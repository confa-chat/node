// Package config provides configuration management for the Confa Hub application
package config

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	nodev1 "github.com/confa-chat/node/src/proto/confa/node/v1"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// AuthProviderOpenIDConnect contains OpenID Connect configuration
type AuthProviderOpenIDConnect struct {
	Issuer       string `koanf:"issuer"`
	ClientID     string `koanf:"clientid"`
	ClientSecret string `koanf:"clientsecret"`
}

// AuthProvider represents an authentication provider configuration
type AuthProvider struct {
	ID            string                    `koanf:"id"`
	Name          string                    `koanf:"name"`
	OpenIDConnect AuthProviderOpenIDConnect `koanf:"openidconnect"`
}

// VoiceRelay represents a voice relay service configuration
type VoiceRelay struct {
	ID      string `koanf:"id"`
	Name    string `koanf:"name"`
	Address string `koanf:"address"`
}

// AttachmentStorage represents configuration for attachment storage
type AttachmentStorage struct {
	// Type is either "local" or "s3"
	Type string `koanf:"type"`

	// Local filesystem storage settings
	Local struct {
		// Path is the directory where attachments will be stored
		Path string `koanf:"path"`
	} `koanf:"local"`

	// S3 storage settings
	S3 struct {
		// Endpoint is the S3 endpoint URL (can be AWS or compatible service)
		Endpoint string `koanf:"endpoint"`
		// Region is the S3 region
		Region string `koanf:"region"`
		// Bucket is the S3 bucket name
		Bucket string `koanf:"bucket"`
		// AccessKeyID is the S3 access key ID
		AccessKeyID string `koanf:"accesskeyid"`
		// SecretAccessKey is the S3 secret access key
		SecretAccessKey string `koanf:"secretaccesskey"`
		// UsePathStyle determines whether to use path-style URLs
		UsePathStyle bool `koanf:"usepathstyle"`
	} `koanf:"s3"`
}

// Config represents the application configuration
type Config struct {
	DB               string            `koanf:"db"`
	AuthProviders    []AuthProvider    `koanf:"authproviders"`
	VoiceRelays      []VoiceRelay      `koanf:"voicerelays"`
	AttachmentConfig AttachmentStorage `koanf:"attachment"`
}

// Load loads configuration from YAML file and environment variables
// Environment variables take precedence over the YAML configuration
func Load(configFile string) (*Config, error) {
	k := koanf.New(".")

	// Load from YAML file if provided
	if configFile != "" {
		if err := k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("error loading config from file %s: %w", configFile, err)
		}
	}

	// Load environment variables (these take precedence over the YAML config)
	err := k.Load(env.Provider("KONFA_HUB_", ".", func(s string) string {
		s = strings.TrimPrefix(s, "KONFA_HUB_")
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "_", ".")
		return s
	}), nil)
	if err != nil {
		return nil, fmt.Errorf("error loading config from environment: %w", err)
	}

	err = parseMapToSlice(k, "authproviders")
	if err != nil {
		log.Printf("warning failed to parse auth providers: %v", err)
	}

	err = parseMapToSlice(k, "voicerelays")
	if err != nil {
		log.Printf("warning failed to parse voice relays: %v", err)
	}

	var cfg Config
	// Unmarshal the config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// Validate the configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	log.Printf("Loaded configuration: %+v\n", cfg)

	return &cfg, nil
}

func parseMapToSlice(k *koanf.Koanf, key string) error {
	m, ok := k.Get(key).(map[string]any)
	if !ok {
		return nil
	}

	s := make([]any, len(m))
	for k, v := range m {
		i, err := strconv.Atoi(k)
		if err != nil {
			return fmt.Errorf("key %s is not an integer: %w", k, err)
		}
		if i > len(s) {
			return fmt.Errorf("key %s is out of range: %d", k, i)
		}
		s[i] = v
	}

	return k.Set(key, s)
}

// validateConfig validates the configuration values
func validateConfig(cfg *Config) error {
	// Check if AuthProviders slice is populated
	if len(cfg.AuthProviders) == 0 {
		return fmt.Errorf("no auth providers configured")
	}

	for _, v := range cfg.AuthProviders {
		if v.ID == "" {
			return fmt.Errorf("auth provider ID is required")
		}
		if v.Name == "" {
			return fmt.Errorf("auth provider name is required")
		}
		if v.OpenIDConnect.Issuer == "" {
			return fmt.Errorf("auth provider issuer is required")
		}
		if v.OpenIDConnect.ClientID == "" {
			return fmt.Errorf("auth provider client ID is required")
		}
		if v.OpenIDConnect.ClientSecret == "" {
			return fmt.Errorf("auth provider client secret is required")
		}
	}

	// If no voice relays are configured, add a default one
	if len(cfg.VoiceRelays) == 0 {
		return fmt.Errorf("no voice relays configured")
	}

	for _, v := range cfg.VoiceRelays {
		if v.ID == "" {
			return fmt.Errorf("voice relay ID is required")
		}
		if v.Name == "" {
			return fmt.Errorf("voice relay name is required")
		}
		if v.Address == "" {
			return fmt.Errorf("voice relay address is required")
		}
	}

	// Validate attachment storage configuration
	if cfg.AttachmentConfig.Type == "" {
		// Default to local storage if not specified
		cfg.AttachmentConfig.Type = "local"
		if cfg.AttachmentConfig.Local.Path == "" {
			cfg.AttachmentConfig.Local.Path = "./attachments"
		}
	}

	switch cfg.AttachmentConfig.Type {
	case "local":
		if cfg.AttachmentConfig.Local.Path == "" {
			return fmt.Errorf("attachment local storage path is required")
		}
	case "s3":
		if cfg.AttachmentConfig.S3.Bucket == "" {
			return fmt.Errorf("attachment S3 bucket is required")
		}
		if cfg.AttachmentConfig.S3.Region == "" {
			return fmt.Errorf("attachment S3 region is required")
		}
		if cfg.AttachmentConfig.S3.AccessKeyID == "" {
			return fmt.Errorf("attachment S3 access key ID is required")
		}
		if cfg.AttachmentConfig.S3.SecretAccessKey == "" {
			return fmt.Errorf("attachment S3 secret access key is required")
		}
	default:
		return fmt.Errorf("invalid attachment storage type: %s", cfg.AttachmentConfig.Type)
	}

	return nil
}

// GetHubAuthProviders converts configuration AuthProviders to hubv1.AuthProvider format
func (c *Config) GetHubAuthProviders() []*nodev1.AuthProvider {
	providers := make([]*nodev1.AuthProvider, 0, len(c.AuthProviders))

	for _, provider := range c.AuthProviders {
		providers = append(providers, &nodev1.AuthProvider{
			Id:   provider.ID,
			Name: provider.Name,
			Protocol: &nodev1.AuthProvider_OpenidConnect{
				OpenidConnect: &nodev1.OpenIDConnect{
					Issuer:       provider.OpenIDConnect.Issuer,
					ClientId:     provider.OpenIDConnect.ClientID,
					ClientSecret: provider.OpenIDConnect.ClientSecret,
				},
			},
		})
	}

	return providers
}

// GetHubVoiceRelays converts configuration VoiceRelays to hubv1.VoiceRelay format
func (c *Config) GetHubVoiceRelays() []*nodev1.VoiceRelay {
	relays := make([]*nodev1.VoiceRelay, 0, len(c.VoiceRelays))

	for _, relay := range c.VoiceRelays {
		relays = append(relays, &nodev1.VoiceRelay{
			Id:      relay.ID,
			Name:    relay.Name,
			Address: relay.Address,
		})
	}

	return relays
}
