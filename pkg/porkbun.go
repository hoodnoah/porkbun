package porkbun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	dnsRetrieveNameTypeURL = "https://api.porkbun.com/api/json/v3/dns/retrieveByNameType/%s/TXT/%s" // append domain name, subdomain (optional)
	dnsCreateURL           = "https://api.porkbun.com/api/json/v3/dns/create/%s"                    // append domain name
	dnsDeleteNameTypeURL   = "https://api.porkbun.com/api/json/v3/dns/deleteByNameType/%s/TXT/%s"   // append domain name, subdomain (optional)
	dnsDeleteByIDURL       = "https://api.porkbun.com/api/json/v3/dns/delete/%s/%s"                 // append domain name, record ID
)

type IDecodable interface {
	Decode(*json.Decoder) error
}

type PorkBun struct {
	httpClient http.Client
	apiKey     string
	secretKey  string
}

type DNSCreateNameTypeArgs struct {
	APIKey    string `json:"apikey"`
	SecretKey string `json:"secretapikey"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	TTL       string `json:"ttl"`
}

type DNSCreateResponse struct {
	Status string      `json:"status"`
	ID     json.Number `json:"id"`
}

func (d *DNSCreateResponse) Decode(decoder *json.Decoder) error {
	return decoder.Decode(d)
}

type DNSDeleteNameTypeArgs struct {
	APIKey    string `json:"apikey"`
	SecretKey string `json:"secretapikey"`
}

type DNSDeleteResponse struct {
	Status string `json:"status"`
}

func (d *DNSDeleteResponse) Decode(decoder *json.Decoder) error {
	return decoder.Decode(d)
}

type DNSRecord struct {
	ID         json.Number `json:"id"`
	Name       string      `json:"name"`
	RecordType string      `json:"type"`
	Content    string      `json:"content"`
	TTL        string      `json:"ttl"`
	Priority   string      `json:"prio"`
	Notes      string      `json:"notes"`
}

type dnsRetrieveArgs struct {
	APIKey    string `json:"apikey"`
	SecretKey string `json:"secretapikey"`
}

type dnsRetrieveResponse struct {
	Status  string      `json:"status"`
	Records []DNSRecord `json:"records"`
}

func (d *dnsRetrieveResponse) Decode(decoder *json.Decoder) error {
	return decoder.Decode(d)
}

// constructor for porkbun struct
func NewPorkbun(apiKey string, secretKey string) *PorkBun {
	return &PorkBun{
		// ACME challenge handling is latency-tolerant (cert-manager waits
		// minutes), so a generous timeout is correct here: a 10s ceiling
		// turns a single transient DNS stall (~5s on clusters with the
		// UDP conntrack race) into a hard failure for no benefit.
		httpClient: http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		secretKey:  secretKey,
	}
}

// setter for the porkbun API Key
func (p *PorkBun) SetAPIKey(apikey string) {
	p.apiKey = apikey
}

// setter for the porkbun Secret Key
func (p *PorkBun) SetSecretKey(secretKey string) {
	p.secretKey = secretKey
}

// Creates a new DNS entry by name and type (subdomain, record type).
// Only TXT records currently supported.
//
// NOTE: this method unconditionally creates the record. Multiple TXT records
// with the same name but different content are valid and REQUIRED for ACME
// DNS01 (e.g. an apex + wildcard certificate produces two challenges at the
// same FQDN). Content-aware deduplication is the caller's responsibility.
// The previous name-only existence check made the second create at a given
// name a silent no-op, which broke exactly that scenario.
func (p *PorkBun) CreateDNSByNameType(domain string, subdomain string, content string) error {
	createUrl := fmt.Sprintf(dnsCreateURL, domain)

	createArgs := DNSCreateNameTypeArgs{
		APIKey:    p.apiKey,
		SecretKey: p.secretKey,
		Name:      subdomain,
		Type:      "TXT",
		Content:   content,
		TTL:       "600",
	}

	createBytes, err := json.Marshal(createArgs)
	if err != nil {
		return err
	}
	bodyBuffer := bytes.NewBuffer(createBytes)

	var createResponse DNSCreateResponse
	if err = p.submitRequest(createUrl, bodyBuffer, &createResponse); err != nil {
		return err
	}

	if strings.ToLower(createResponse.Status) != "success" {
		return fmt.Errorf("failed to create record with status %s", createResponse.Status)
	}

	return nil
}

// Deletes a DNS entry by name and type (subdomain, record type).
// Only TXT records currently supported.
//
// WARNING: Porkbun's deleteByNameType endpoint deletes ALL records with the
// given name and type, regardless of content. To delete a single specific
// record, retrieve records with RetrieveDNSByNameType, match on content,
// and use DeleteDNSByID with the matched record's ID.
func (p *PorkBun) DeleteDNSByNameType(domain string, subdomain string) error {
	// check if the record exists
	recordExists, err := p.recordExists(domain, subdomain)
	if err != nil {
		return err
	}

	if !recordExists {
		return nil
	}

	// record exists, delete it
	deleteUrl := fmt.Sprintf(dnsDeleteNameTypeURL, domain, subdomain)
	deleteArgs := DNSDeleteNameTypeArgs{
		APIKey:    p.apiKey,
		SecretKey: p.secretKey,
	}
	bodyBytes, err := json.Marshal(deleteArgs)
	if err != nil {
		return err
	}
	bodyBuffer := bytes.NewBuffer(bodyBytes)

	var deleteResponse DNSDeleteResponse
	if err = p.submitRequest(deleteUrl, bodyBuffer, &deleteResponse); err != nil {
		return err
	}

	if strings.ToLower(deleteResponse.Status) != "success" {
		return fmt.Errorf("failed to submit delete request with status %s", deleteResponse.Status)
	}

	return nil
}

// Deletes a single DNS record by its Porkbun record ID.
// Unlike DeleteDNSByNameType, this affects exactly one record, making it
// safe when multiple records share a name (e.g. concurrent ACME challenges).
func (p *PorkBun) DeleteDNSByID(domain string, id string) error {
	deleteUrl := fmt.Sprintf(dnsDeleteByIDURL, domain, id)
	deleteArgs := DNSDeleteNameTypeArgs{
		APIKey:    p.apiKey,
		SecretKey: p.secretKey,
	}
	bodyBytes, err := json.Marshal(deleteArgs)
	if err != nil {
		return err
	}
	bodyBuffer := bytes.NewBuffer(bodyBytes)

	var deleteResponse DNSDeleteResponse
	if err = p.submitRequest(deleteUrl, bodyBuffer, &deleteResponse); err != nil {
		return err
	}

	if strings.ToLower(deleteResponse.Status) != "success" {
		return fmt.Errorf("failed to submit delete-by-id request with status %s", deleteResponse.Status)
	}

	return nil
}

// Fetches and returns a list of DNS records by domain, subdomain
// only TXT records supported.
// Returns an empty list in the event there are no records for the given domain
func (p *PorkBun) RetrieveDNSByNameType(domain string, subdomain string) ([]DNSRecord, error) {
	url := fmt.Sprintf(dnsRetrieveNameTypeURL, domain, subdomain)

	// create request body
	args := dnsRetrieveArgs{
		APIKey:    p.apiKey,
		SecretKey: p.secretKey,
	}
	bodyBytes, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	bodyBuffer := bytes.NewBuffer(bodyBytes)

	// submit the request
	var response dnsRetrieveResponse
	err = p.submitRequest(url, bodyBuffer, &response)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(response.Status) != "success" {
		return nil, fmt.Errorf("failed to retrieve DNS records with status %s", response.Status)
	}

	return response.Records, nil
}

// determines if a record exists; true if it does, false if it doesn't
func (p *PorkBun) recordExists(domain, subdomain string) (bool, error) {
	url := fmt.Sprintf(dnsRetrieveNameTypeURL, domain, subdomain)

	// create request body
	args := dnsRetrieveArgs{
		APIKey:    p.apiKey,
		SecretKey: p.secretKey,
	}

	// jsonify the request body
	body, err := json.Marshal(args)
	if err != nil {
		return false, err
	}
	bodyBuffer := bytes.NewBuffer(body)

	// submit request
	var response dnsRetrieveResponse
	err = p.submitRequest(url, bodyBuffer, &response)
	if err != nil {
		return false, err
	}

	if strings.ToLower(response.Status) != "success" {
		return false, fmt.Errorf("unexpected status received from API: %s", response.Status)
	}

	// if records array contains records, infer that the record exists
	if len(response.Records) != 0 {
		return true, nil
	}

	return false, nil
}

// submit a generic request to the PorkBun API
// decoding the result into the provided return struct, implementing IDecodable interface
func (p *PorkBun) submitRequest(url string, bodyBuffer *bytes.Buffer, returnStruct IDecodable) error {
	req, err := http.NewRequest("POST", url, bodyBuffer)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	// close the body once handled; without this, every API call leaks the
	// connection, which matters in a long-running webhook process
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %s, responseL %s", resp.Status, string(responseBody))
	}

	decoder := json.NewDecoder(resp.Body)

	return returnStruct.Decode(decoder)
}
