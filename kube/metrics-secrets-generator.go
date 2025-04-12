package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AdminUser            string `json:"adminUser"`
	AdminPassword        string `json:"adminPassword"`
	AWSAccessKey         string `json:"AWS_ACCESS_KEY"`
	AWSSecretAccessKey   string `json:"AWS_SECRET_ACCESS_KEY"`
}

func prompt(label string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", label)
	value, _ := reader.ReadString('\n')
	return strings.TrimSpace(value)
}

func encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func readConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	err = json.NewDecoder(file).Decode(config)
	return config, err
}

func readLicenseFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func createSecretYAML(name string, data map[string]string) string {
	yaml := fmt.Sprintf("apiVersion: v1\nkind: Secret\nmetadata:\n  name: %s\ntype: Opaque\ndata:\n", name)
	for key, val := range data {
		yaml += fmt.Sprintf("  %s: %s\n", key, encode(val))
	}
	return yaml
}

func writeToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

func main() {
	// Flags
	adminUserFlag := flag.String("adminUser", "", "Admin username")
	adminPassFlag := flag.String("adminPassword", "", "Admin password")
	licenseFileFlag := flag.String("licensefile", "", "Path to license.jwt file")
	awsKeyFlag := flag.String("awsKey", "", "AWS Access Key")
	awsSecretFlag := flag.String("awsSecret", "", "AWS Secret Access Key")
	configPath := "config.json"
	flag.Parse()

	var config *Config
	var err error

	// Load config.json if present
	if _, err := os.Stat(configPath); err == nil {
		config, err = readConfig(configPath)
		if err != nil {
			fmt.Println("Error reading config.json:", err)
			return
		}
	} else {
		config = &Config{}
	}

	// Admin credentials
	adminUser := *adminUserFlag
	if adminUser == "" {
		if config.AdminUser != "" {
			adminUser = config.AdminUser
		} else {
			adminUser = prompt("Enter adminUser")
		}
	}

	adminPass := *adminPassFlag
	if adminPass == "" {
		if config.AdminPassword != "" {
			adminPass = config.AdminPassword
		} else {
			adminPass = prompt("Enter adminPassword")
		}
	}

	// AWS credentials
	awsKey := *awsKeyFlag
	if awsKey == "" {
		if config.AWSAccessKey != "" {
			awsKey = config.AWSAccessKey
		} else {
			awsKey = prompt("Enter AWS_ACCESS_KEY")
		}
	}

	awsSecret := *awsSecretFlag
	if awsSecret == "" {
		if config.AWSSecretAccessKey != "" {
			awsSecret = config.AWSSecretAccessKey
		} else {
			awsSecret = prompt("Enter AWS_SECRET_ACCESS_KEY")
		}
	}

	// License file
	licenseData := ""
	if *licenseFileFlag != "" {
		licenseData, err = readLicenseFile(*licenseFileFlag)
		if err != nil {
			fmt.Println("Error reading license file:", err)
			return
		}
	} else {
		licenseData = prompt("Enter contents of license.jwt")
	}

	// Create secrets
	metricsAdminSecret := createSecretYAML("metrics-admin-secret", map[string]string{
		"adminUser":     adminUser,
		"adminPassword": adminPass,
	})

	createLicenseSecret := createSecretYAML("create-license-secret", map[string]string{
		"license.jwt": licenseData,
	})

	metricsBucketSecret := createSecretYAML("metrics-bucket-secret", map[string]string{
		"AWS_ACCESS_KEY":        awsKey,
		"AWS_SECRET_ACCESS_KEY": awsSecret,
	})

	// Write to files
	if err := writeToFile("metrics-admin-secret.yaml", metricsAdminSecret); err != nil {
		fmt.Println("Failed to write metrics-admin-secret.yaml:", err)
	}
	if err := writeToFile("metrics-license-secret.yaml", createLicenseSecret); err != nil {
		fmt.Println("Failed to write create-license-secret.yaml:", err)
	}
	if err := writeToFile("metrics-bucket-secret.yaml", metricsBucketSecret); err != nil {
		fmt.Println("Failed to write metrics-bucket-secret.yaml:", err)
	}

	fmt.Println("✅ Kubernetes Secret manifests written to:")
	fmt.Println("  - metrics-admin-secret.yaml")
	fmt.Println("  - create-license-secret.yaml")
	fmt.Println("  - metrics-bucket-secret.yaml")
}
