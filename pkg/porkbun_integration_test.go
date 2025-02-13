package porkbun_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	porkbun "github.com/hoodnoah/porkbun/pkg"
)

func TestIntegration_CreateAndDeleteDNSRecord(t *testing.T) {
	apiKey := os.Getenv("PORKBUN_API_KEY")
	secretKey := os.Getenv("PORKBUN_SECRET_KEY")
	testDomain := os.Getenv("PORKBUN_TEST_DOMAIN")

	if apiKey == "" {
		t.Skip("Skipping integration test: required environment variable PORKBUN_API_KEY not set.")
	}

	if secretKey == "" {
		t.Skip("Skipping integration test: required environment variable PORKBUN_SECRET_KEY not set.")
	}

	if testDomain == "" {
		t.Skip("Skipping integration test: required environment variable PORKBUN_TEST_DOMAIN not set.")
	}

	client := porkbun.NewPorkbun(apiKey, secretKey)

	testTimeStamp := time.Now().UnixNano()

	subdomain := fmt.Sprintf("test-%d", testTimeStamp)
	content := fmt.Sprintf("integration-test-content-%d", testTimeStamp)

	// clean-up before testing
	t.Cleanup(func() {
		_ = client.DeleteDNSByNameType(testDomain, subdomain)
	})

	// test creation
	if err := client.CreateDNSByNameType(testDomain, subdomain, content); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify creation
	records, err := client.RetrieveDNSByNameType(testDomain, subdomain)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	// expect a record
	if len(records) == 0 {
		t.Fatal("No records found after creation.")
	}

	found := false
	for _, record := range records {
		if record.Content == content {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Failed to find expected content in retrieved records.")
	}

	// Test deletion
	if err := client.DeleteDNSByNameType(testDomain, subdomain); err != nil {
		t.Fatalf("Failed to delete created record: %v", err)
	}

	// Verify deletion
	records, err = client.RetrieveDNSByNameType(testDomain, subdomain)
	if err != nil {
		t.Fatalf("Failed to retrieve list of records after deletion: %v", err)
	}

	if len(records) > 0 {
		t.Fatalf("expected to receive no records, received %d", len(records))
	}
}
