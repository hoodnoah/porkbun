# 🚧 WORK IN PROGRESS 🚧

This is a work in progress. The API is extremely likely to change significantly. This is not yet suitable for
any production workload. For anyone using this at this stage of development, pin your version number to a specific tag that you have verified works as expected.

# Porkbun

## Scope

This is a simple Go wrapper around the [porkbun web api](https://porkbun.com/api/json/v3/documentation).

[Porkbun.com](https://porkbun.com/) is a simple to use, inexpensive domain registrar.

Their API exposes many endpoints for the management of domains. Those currently implemented are below.

Porkbun exposes their API to users who have enabled API access on a given domain. API access is afforded through two keys provided to the user: the API Key, and the Secret Key. Both keys are required for action against the API.

## Usage

All API operations are conducted through the use of a `Porkbun` struct. The constructor for the porkbun struct requires two arguments; the `api key` and `secret api key` as provided by Porkbun.

Example:

```go
import (
  "fmt"
  "os"

  porkbun "github.com/hoodnoah/porkbun/pkg"
)

func main() {
  apiKey := os.Getenv("PORKBUN_API_KEY")
  secretKey := os.Getenv("PORKBUN_SECRET_KEY")
  domain := os.Getenv("PORKBUN_DOMAIN")

  // test, assert that environment variables are set and retrieved correctly
  // ...

  // initialize the client
  pbClient := porkbun.NewPorkbun(apiKey, secretKey)

  // create a DNS record; as of v0.0.1, only TXT records are supported.
  subdomain := os.Getenv("PORKBUN_SUBDOMAIN") // pick a subdomain; leave empty for a root domain, or use * for a wildcard

  content := "record_content" // assign whatever content you need the record to return when retrieved

  if err := client.CreateDNSByNameType(domain, subdomain, content); err != nil {
    fmt.Printf("failed to create record: %v", err)
  }

  // retrieve existing records; as of v0.0.1, only TXT records are supported.
  records, err := client.RetrieveDNSByNameType(domain, subdomain)
  if err != nil {
    fmt.Printf("failed to retrieve records: %v", err)
  }
  // if successful, records is a list of records following the format of the documentation:
  // https://porkbun.com/api/json/v3/documentation#DNS%20Retrieve%20Records%20by%20Domain,%20Subdomain%20and%20Type

  // delete records; as of v0.0.1, only TXT records are supported.
  if err := client.DeleteDNSByNameType(domain, subdomain); err != nil {
    fmt.Printf("failed to delete records: %v", err)
  }
}
```

## Implemented

### DNS

#### [Create by Name Type](https://porkbun.com/api/json/v3/documentation#DNS%20Create%20Record)

- Allows the user to create a DNS record for a given domain and subdomain.
- As of v0.0.1, only TXT records are implemented.

#### [Delete by Name Type](https://porkbun.com/api/json/v3/documentation#DNS%20Delete%20Records%20by%20Domain,%20Subdomain%20and%20Type)

- Allows the user to delete DNS records by domain, subdomain, and type.
- As of v0.0.1, only TXT records are implemented.

#### [Retrieve by Name Type](https://porkbun.com/api/json/v3/documentation#DNS%20Retrieve%20Records%20by%20Domain,%20Subdomain%20and%20Type)

- Allows the user to retrieve DNS records by domain, subdomain, and type
- As of v0.0.1, only TXT records are implemented.
