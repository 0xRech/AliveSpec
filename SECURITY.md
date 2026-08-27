# Security Policy

AliveSpec is designed to observe **operational metadata**, not application content.

## Data that must not be captured

AliveSpec should not record:

- packet payloads
- passwords or tokens
- private keys
- environment-variable values
- file contents
- request or response bodies

File requirements may contain a SHA-256 fingerprint, but not file contents.

## Reporting a vulnerability

Please use a private GitHub security advisory instead of a public issue when reporting a security vulnerability.
