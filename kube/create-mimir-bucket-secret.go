package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"text/template"
)

const secretTemplate = `apiVersion: v1
kind: Secret
metadata:
  name: mimir-bucket-secret
type: Opaque
data:
  AWS_ACCESS_KEY_ID: {{ .AccessKeyID }}
  AWS_SECRET_ACCESS_KEY: {{ .SecretAccessKey }}
`

type SecretData struct {
	AccessKeyID     string
	SecretAccessKey string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <AWS_ACCESS_KEY_ID> <AWS_SECRET_ACCESS_KEY>\n", os.Args[0])
		os.Exit(1)
	}

	accessKeyID := os.Args[1]
	secretAccessKey := os.Args[2]

	encoded := SecretData{
		AccessKeyID:     base64.StdEncoding.EncodeToString([]byte(accessKeyID)),
		SecretAccessKey: base64.StdEncoding.EncodeToString([]byte(secretAccessKey)),
	}

	tmpl, err := template.New("secret").Parse(secretTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse template: %v\n", err)
		os.Exit(1)
	}

	file, err := os.Create("mimir-bucket-secret.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	err = tmpl.Execute(file, encoded)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write secret manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Secret manifest written to mimir-bucket-secret.yaml")
}
