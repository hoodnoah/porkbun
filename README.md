🚧 WORK IN PROGRESS 🚧

# Porkbun

## Scope

This is a simple Go wrapper around the [porkbun web api](https://porkbun.com/api/json/v3/documentation).

[Porkbun.com](https://porkbun.com/) is a simple to use, inexpensive domain registrar.

Their API exposes many endpoints for the management of domains. Those currently implemented are below.

Porkbun exposes their API to users who have enabled API access on a given domain. API access is afforded through two keys provided to the user: the API Key, and the Secret Key. Both keys are required for action against the API.

## Implemented

### DNS

#### Create

- Allows the user to create a DNS record for a given domain.
- Currently, only TXT records are implemented.
