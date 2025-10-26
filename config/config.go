package config

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AwsProfile string
	TabKey     int
}

// HibiscusConfig wraps the persistent config in a top-level 'hibiscus' key
type HibiscusConfig struct {
	Hibiscus PersistentConfig `yaml:"hibiscus"`
}

// PersistentConfig only stores values we want to persist between sessions
type PersistentConfig struct {
	ServiceName string `yaml:"service_name"`
}

var (
	globalConfig *Config = nil
	configMutex  sync.RWMutex
	configDir    = getConfigDir()
	configFile   = filepath.Join(configDir, "config.yaml")
)

// getConfigDir returns the appropriate config directory path
func getConfigDir() string {
	// Check if XDG_CONFIG_HOME is set
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "hibiscus")
	}
	// Fall back to $HOME/.config
	return filepath.Join(os.Getenv("HOME"), ".config", "hibiscus")
}

// tabKeyToServiceName converts a tab key integer to a service name string
func tabKeyToServiceName(tabKey int) string {
	switch tabKey {
	case ECR_TAB:
		return "ecr"
	case ROUTE53_TAB:
		return "route53"
	case ELB_TAB:
		return "elb"
	default:
		return "ecr" // Default to ECR if unknown
	}
}

// serviceNameToTabKey converts a service name string to a tab key integer
func serviceNameToTabKey(serviceName string) int {
	switch serviceName {
	case "ecr":
		return ECR_TAB
	case "route53":
		return ROUTE53_TAB
	case "elb":
		return ELB_TAB
	default:
		return ECR_TAB // Default to ECR if unknown
	}
}

func Initialize() *Config {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Create default config
	globalConfig = &Config{
		AwsProfile: "default",
		TabKey:     ECR_TAB, // Default tab
	}

	// TODO: Try to load saved tab from file
	// hibiscusConfig, err := loadConfigFromFile()
	// if err == nil && hibiscusConfig != nil && hibiscusConfig.Hibiscus.ServiceName != "" {
	// 	// Only update the tab key, not the AWS profile
	// 	globalConfig.TabKey = serviceNameToTabKey(hibiscusConfig.Hibiscus.ServiceName)
	// }

	return globalConfig
}

func GetConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		Initialize()
	}

	return globalConfig
}

func SetAwsProfile(profile string) {
	configMutex.Lock()
	defer configMutex.Unlock()

	globalConfig.AwsProfile = profile
	// Don't save the profile - it's not persisted
}

func SetTabKey(key int) {
	configMutex.Lock()
	defer configMutex.Unlock()

	globalConfig.TabKey = key

	// Save tab key when it changes
	saveConfigToFile()
}

// Load configuration from file
// WARN: dead_code
func loadConfigFromFile() (*HibiscusConfig, error) {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config HibiscusConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save configuration to file
func saveConfigToFile() error {
	// Create a persistent config with the values we want to persist
	persistentConfig := PersistentConfig{
		ServiceName: tabKeyToServiceName(globalConfig.TabKey),
	}

	// Wrap it in the top-level HibiscusConfig
	hibiscusConfig := HibiscusConfig{
		Hibiscus: persistentConfig,
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	// Create an encoder with 2-space indentation
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	// Encode the config
	if err := encoder.Encode(hibiscusConfig); err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(configFile, buf.Bytes(), 0o644)
}
