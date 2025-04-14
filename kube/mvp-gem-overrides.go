package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Input from config or CLI
type InputConfig struct {
	AdminBucket  string `json:"admin_bucket"`
	RulerBucket  string `json:"ruler_bucket"`
	BlocksBucket string `json:"blocks_bucket"`
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	Endpoint     string `json:"endpoint"`
}

// YAML Structs
type SecretRef struct {
	Name string `yaml:"name"`
}
type EnvFrom struct {
	SecretRef SecretRef `yaml:"secretRef"`
}
type S3Config struct {
	BucketName      string `yaml:"bucket_name"`
	AccessKeyID     string `yaml:"access_key_id"`
	Endpoint        string `yaml:"endpoint"`
	SecretAccessKey string `yaml:"secret_access_key"`
}
type AdminStorage struct {
	Storage struct {
		S3 S3Config `yaml:"s3"`
	} `yaml:"storage"`
}
type MimirConfig struct {
	AdminClient        AdminStorage `yaml:"admin_client"`
	AlertmanagerStorage struct {
		S3 S3Config `yaml:"s3"`
	} `yaml:"alertmanager_storage"`
	BlocksStorage struct {
		Backend string   `yaml:"backend"`
		S3      S3Config `yaml:"s3"`
	} `yaml:"blocks_storage"`
	RulerStorage struct {
		S3 S3Config `yaml:"s3"`
	} `yaml:"ruler_storage"`
}
type Values struct {
	Global struct {
		ExtraEnvFrom   []EnvFrom          `yaml:"extraEnvFrom"`
		PodAnnotations map[string]string `yaml:"podAnnotations"`
	} `yaml:"global"`
	Minio struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"minio"`
	Mimir struct {
		StructuredConfig MimirConfig `yaml:"structuredConfig"`
	} `yaml:"mimir"`
}

// Constants
const outputFile = "gem-overrides.yaml"

const sampleJSON = `
Sample gem-overrides.json:

{
  "admin_bucket": "my-admin-bucket",
  "ruler_bucket": "my-ruler-bucket",
  "blocks_bucket": "my-blocks-bucket",
  "access_key": "${AWS_ACCESS_KEY_ID}",
  "secret_key": "${AWS_SECRET_ACCESS_KEY}",
  "endpoint": "s3.amazonaws.com"
}
`

// Read config file
func readConfigFile(filename string) (InputConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return InputConfig{}, err
	}
	var cfg InputConfig
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

// Validate input
func validateInput(cfg InputConfig) error {
	if cfg.AdminBucket == "" || cfg.RulerBucket == "" || cfg.BlocksBucket == "" ||
		cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Endpoint == "" {
		return fmt.Errorf("one or more required fields are missing")
	}
	return nil
}

func main() {
	// CLI flags
	configFile := flag.String("config", "gem-overrides.json", "Path to gem-overrides.json")
	adminBucket := flag.String("admin-bucket", "", "Admin S3 bucket")
	rulerBucket := flag.String("ruler-bucket", "", "Ruler S3 bucket")
	blocksBucket := flag.String("blocks-bucket", "", "Blocks S3 bucket")
	accessKey := flag.String("access-key", "", "AWS access key")
	secretKey := flag.String("secret-key", "", "AWS secret key")
	endpoint := flag.String("endpoint", "", "S3 endpoint")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Usage:
  go run main.go [--config FILE] [--admin-bucket NAME] ...

Description:
  Generates Helm overrides YAML for Grafana Enterprise Metrics.

Flags:
`)
		flag.PrintDefaults()
		fmt.Println(sampleJSON)
	}

	flag.Parse()

	var cfg InputConfig
	var configLoaded bool

	// Load JSON config if available
	if fileCfg, err := readConfigFile(*configFile); err == nil {
		cfg = fileCfg
		configLoaded = true
	}

	// Override with CLI flags
	if *adminBucket != "" {
		cfg.AdminBucket = *adminBucket
		configLoaded = true
	}
	if *rulerBucket != "" {
		cfg.RulerBucket = *rulerBucket
		configLoaded = true
	}
	if *blocksBucket != "" {
		cfg.BlocksBucket = *blocksBucket
		configLoaded = true
	}
	if *accessKey != "" {
		cfg.AccessKey = *accessKey
		configLoaded = true
	}
	if *secretKey != "" {
		cfg.SecretKey = *secretKey
		configLoaded = true
	}
	if *endpoint != "" {
		cfg.Endpoint = *endpoint
		configLoaded = true
	}

	// If nothing loaded, show help and exit
	if !configLoaded {
		fmt.Println("❌ No config file found and no CLI flags provided.")
		flag.Usage()
		os.Exit(1)
	}

	// Validate inputs
	if err := validateInput(cfg); err != nil {
		fmt.Printf("❌ Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Create Helm YAML structure
	var values Values
	values.Global.ExtraEnvFrom = []EnvFrom{{SecretRef: SecretRef{Name: "metrics-bucket-secret"}}}
	values.Global.PodAnnotations = map[string]string{"bucketSecretVersion": "0"}
	values.Minio.Enabled = false

	s3Admin := S3Config{cfg.AdminBucket, cfg.AccessKey, cfg.Endpoint, cfg.SecretKey}
	s3Ruler := S3Config{cfg.RulerBucket, cfg.AccessKey, cfg.Endpoint, cfg.SecretKey}
	s3Blocks := S3Config{cfg.BlocksBucket, cfg.AccessKey, cfg.Endpoint, cfg.SecretKey}

	values.Mimir.StructuredConfig.AdminClient.Storage.S3 = s3Admin
	values.Mimir.StructuredConfig.AlertmanagerStorage.S3 = s3Ruler
	values.Mimir.StructuredConfig.RulerStorage.S3 = s3Ruler
	values.Mimir.StructuredConfig.BlocksStorage.Backend = "s3"
	values.Mimir.StructuredConfig.BlocksStorage.S3 = s3Blocks

	// Marshal to YAML
	yamlData, err := yaml.Marshal(values)
	if err != nil {
		fmt.Printf("❌ Failed to marshal YAML: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile(outputFile, yamlData, 0644)
	if err != nil {
		fmt.Printf("❌ Failed to write %s: %v\n", outputFile, err)
		os.Exit(1)
	}

	fmt.Printf("✅ %s created successfully.\n", outputFile)
}
