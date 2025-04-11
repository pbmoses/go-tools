package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
)

func main() {
	// Define command-line flags
	namespace := flag.String("namespace", "", "Kubernetes namespace (required)")
	secretName := flag.String("name", "", "Kubernetes secret name (required)")
	licensePath := flag.String("file", "", "Path to license.jwt file (required)")

	flag.Parse()

	// Validate required flags
	if *namespace == "" || *secretName == "" || *licensePath == "" {
		fmt.Println("Error: All flags --namespace, --name, and --file are required.")
		flag.Usage()
		os.Exit(1)
	}

	// Check if file exists
	fileContent, err := os.ReadFile(*licensePath)
	if err != nil {
		fmt.Printf("Error reading file '%s': %v\n", *licensePath, err)
		os.Exit(1)
	}

	// Base64 encode the file content
	encodedContent := base64.StdEncoding.EncodeToString(fileContent)

	// Generate the manifest
	manifest := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
data:
  license.jwt: %s
`, *secretName, *namespace, encodedContent)

	// Write manifest to file
	outputFile := fmt.Sprintf("%s.yaml", *secretName)
	err = os.WriteFile(outputFile, []byte(manifest), 0644)
	if err != nil {
		fmt.Printf("Error writing manifest to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Secret manifest written to %s\n", outputFile)
}

