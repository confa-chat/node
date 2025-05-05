// Package config provides configuration management for the Konfa Hub application
package config

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	hubv1 "github.com/konfa-chat/hub/src/proto/konfa/hub/v1"
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

// Config represents the application configuration
type Config struct {
	DB            string         `koanf:"db"`
	AuthProviders []AuthProvider `koanf:"authproviders"`
	VoiceRelays   []VoiceRelay   `koanf:"voicerelays"`
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

	return nil
}

// GetHubAuthProviders converts configuration AuthProviders to hubv1.AuthProvider format
func (c *Config) GetHubAuthProviders() []*hubv1.AuthProvider {
	providers := make([]*hubv1.AuthProvider, 0, len(c.AuthProviders))

	for _, provider := range c.AuthProviders {
		providers = append(providers, &hubv1.AuthProvider{
			Id:   provider.ID,
			Name: provider.Name,
			Protocol: &hubv1.AuthProvider_OpenidConnect{
				OpenidConnect: &hubv1.OpenIDConnect{
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
func (c *Config) GetHubVoiceRelays() []*hubv1.VoiceRelay {
	relays := make([]*hubv1.VoiceRelay, 0, len(c.VoiceRelays))

	for _, relay := range c.VoiceRelays {
		relays = append(relays, &hubv1.VoiceRelay{
			Id:      relay.ID,
			Name:    relay.Name,
			Address: relay.Address,
		})
	}

	return relays
}
