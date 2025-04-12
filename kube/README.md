## Requirements

Basic understanding of Go, Go properly installed.

## Metrics secrets generator
Thie Go tool will output 3 secrets; one admin, one for S3 storage and one for the license.jwt that is provided. COmmand line arguments can be used or config.json. 
` go build metrics-secrets-generator.go` or `go run  go build metrics-secrets-generator.go`

## Admin Secret Generator

This Go based too will prompt the user for a namespace, secret name, adminUser, adminPassword, Base64-encode the data and create a local file saved as the secret name. 

`go run create-admin-secret.go` or `go build create-admin-secret.go`
## License Secret Generator
`create-license-secret.go` will create a secret containing a license.jwt file. The user will provide the namespace, secret name, and path to the license file as command-line arguments. The license file is base64 encoded, validation takes place for inputs, the secret is saved to a yaml manifest names <secret-name>.yaml

To create the secret `go run create-license-secret.go` or `go build create-license-secret.go`.
Example with the binary built: `./create-license-secret -file license.jwt -name demo-license -namespace demo`create-license-secret.go`

Successful creation: `✅ Secret manifest written to demo-license.yaml`

