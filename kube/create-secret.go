package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func prompt(reader *bufio.Reader, label string) string {
	fmt.Print(label)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Get user input
	namespace := prompt(reader, "Enter the Kubernetes namespace: ")
	secretName := prompt(reader, "Enter the secret name: ")
	adminUser := prompt(reader, "Enter admin username: ")
	adminPassword := prompt(reader, "Enter admin password: ")

	// Base64 encode the user and password
	encodedUser := base64.StdEncoding.EncodeToString([]byte(adminUser))
	encodedPassword := base64.StdEncoding.EncodeToString([]byte(adminPassword))

	// Create manifest content
	manifest := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
data:
  adminUser: %s
  adminPassword: %s
`, secretName, namespace, encodedUser, encodedPassword)

	// Write to file
	filename := fmt.Sprintf("%s.yaml", secretName)
	err := os.WriteFile(filename, []byte(manifest), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Printf("Secret manifest written to %s\n", filename)
}
