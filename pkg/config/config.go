package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/emilkje/cwc/pkg/errors"

	"github.com/sashabaranov/go-openai"
)

var Cfg Config

const (
	configFileName = "cwc.json" // The name of the config file we want to save
)

func NewFromConfigFile() (openai.ClientConfig, error) {
	Cfg, err := LoadConfig()
	if err != nil {
		return openai.ClientConfig{}, err
	}

	// validate the configuration
	err = ValidateConfig(Cfg)
	if err != nil {
		return openai.ClientConfig{}, err
	}

	var config openai.ClientConfig
	if Cfg.Provider == "azure" {
		config = openai.DefaultAzureConfig(Cfg.APIKey(), Cfg.Endpoint)
		config.APIVersion = Cfg.ApiVersion
		config.AzureModelMapperFunc = func(model string) string {
			return Cfg.ModelDeployment
		}
	}
	if Cfg.Provider == "openai" {
		config = openai.DefaultConfig(Cfg.apiKey)
		config.BaseURL = Cfg.Endpoint + "/" + Cfg.ApiVersion
	}

	return config, nil
}

// SanitizeInput trims whitespaces and newlines from a string.
func SanitizeInput(input string) string {
	return strings.TrimSpace(input)
}

type Config struct {
	Provider        string `json:"provider"`
	Endpoint        string `json:"endpoint"`
	ApiVersion      string `json:"apiVersion"`
	ModelDeployment string `json:"modelDeployment"`
	Model           string `json:"model"`
	// Keep APIKey unexported to avoid accidental exposure
	apiKey string
}

// NewConfig creates a new Config object
func NewConfig(provider, endpoint, apiVersion, modelDeployment, model string) *Config {
	return &Config{
		Provider:        provider,
		Endpoint:        endpoint,
		ApiVersion:      apiVersion,
		ModelDeployment: modelDeployment,
		Model:           model,
	}
}

// SetAPIKey sets the confidential field apiKey
func (c *Config) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

// APIKey returns the confidential field apiKey
func (c *Config) APIKey() string {
	return c.apiKey
}

var SupportedProviders = []string{
	"azure",
	"openai",
}

// ValidateConfig checks if a Config object has valid data.
func ValidateConfig(c *Config) error {

	var validationErrors []string
	if c.Provider == "" {
		validationErrors = append(validationErrors, "provider must be provided and not be empty")
	}

	supportedProviderFound := false
	for _, supportedProvider := range SupportedProviders {
		if c.Provider == supportedProvider {
			supportedProviderFound = true
		}
	}
	if !supportedProviderFound {
		validationErrors = append(validationErrors, "provider not supported")
	}
	if c.APIKey() == "" {
		validationErrors = append(validationErrors, "apiKey must be provided and not be empty")
	}

	if c.Endpoint == "" {
		validationErrors = append(validationErrors, "endpoint must be provided and not be empty")
	}

	if c.ApiVersion == "" {
		validationErrors = append(validationErrors, "apiVersion must be provided and not be empty")
	}

	if c.Provider == "azure" {
		if c.ModelDeployment == "" {
			validationErrors = append(validationErrors, "modelDeployment must be provided and not be empty")
		}
	}

	if c.Provider == "openai" {
		if c.Model == "" {
			validationErrors = append(validationErrors, "model must be provided and not be empty")
		}
	}

	if len(validationErrors) > 0 {
		return &errors.ConfigValidationError{Errors: validationErrors}
	}

	return nil
}

// SaveConfig writes the configuration to disk, and the API key to the keyring.
func SaveConfig(config *Config) error {

	// validate the configuration
	err := ValidateConfig(config)
	if err != nil {
		return err
	}

	configDir, err := xdgConfigPath()
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(configDir, configFileName)

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = storeApiKeyInKeyring(config.APIKey())

	if err != nil {
		return err
	}

	return os.WriteFile(configFilePath, data, 0644)
}

// LoadConfig reads the configuration from disk and loads the API key from the keyring.
func LoadConfig() (*Config, error) {
	// Read data from file or secure store
	configDir, err := xdgConfigPath()
	if err != nil {
		return nil, err
	}

	configFilePath := filepath.Join(configDir, configFileName)
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &Cfg)
	if err != nil {
		return nil, err
	}

	apiKey, err := getApiKeyFromKeyring()
	if err != nil {
		return nil, err
	}

	Cfg.SetAPIKey(apiKey)

	return &Cfg, nil
}

func ClearConfig() error {
	configDir, err := xdgConfigPath()
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(configDir, configFileName)

	err = os.Remove(configFilePath)
	if err != nil {
		return err
	}

	err = clearApiKeyInKeyring()
	if err != nil {
		return err
	}

	return nil
}
