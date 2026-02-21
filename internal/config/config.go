package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	configDirName  = ".confluence"
	configFileName = "config"
	configFileType = "yaml"

	defaultFormat = "pretty"
)

// Config holds all configuration for the Confluence CLI.
type Config struct {
	BaseURL string
	Email   string
	Token   string
	Space   string
	Format  string
}

// DefaultConfigPath returns the default path to the configuration file:
// ~/.confluence/config.yaml
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fall back to a relative path when the home directory is unavailable.
		return filepath.Join(configDirName, configFileName+".yaml")
	}
	return filepath.Join(home, configDirName, configFileName+".yaml")
}

// Load reads configuration from the config file and environment variables.
// Priority: env vars > config file > defaults.
//
// If cfgFile is non-empty it is used as the config file path; otherwise the
// default path (~/.confluence/config.yaml) is used.
//
// CLI flag overrides must be applied by the caller after Load() returns.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("format", defaultFormat)

	// Environment variable bindings.
	// We do NOT use AutomaticEnv to avoid accidentally picking up unrelated
	// env vars; instead we bind each key explicitly.
	_ = v.BindEnv("base_url", "CONFLUENCE_BASE_URL")
	_ = v.BindEnv("email", "CONFLUENCE_EMAIL")
	_ = v.BindEnv("token", "CONFLUENCE_TOKEN")
	_ = v.BindEnv("space", "CONFLUENCE_SPACE")

	// Config file location.
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		v.SetConfigType(configFileType)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("determine home directory: %w", err)
		}
		v.SetConfigName(configFileName)
		v.SetConfigType(configFileType)
		v.AddConfigPath(filepath.Join(home, configDirName))
	}

	if err := v.ReadInConfig(); err != nil {
		// When an explicit file path was given, any read error is fatal.
		if cfgFile != "" {
			return nil, fmt.Errorf("read config file %s: %w", cfgFile, err)
		}
		// When using the default search path, a missing file is acceptable.
		// Config file is optional — env vars may provide all values.
	}

	cfg := &Config{
		BaseURL: v.GetString("base_url"),
		Email:   v.GetString("email"),
		Token:   v.GetString("token"),
		Space:   v.GetString("space"),
		Format:  v.GetString("format"),
	}

	if cfg.Format == "" {
		cfg.Format = defaultFormat
	}

	// Strip trailing slashes from BaseURL for consistent URL construction.
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	return cfg, nil
}

// Validate checks that the minimum required fields are present.
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return errors.New("base_url is required (set CONFLUENCE_BASE_URL or base_url in ~/.confluence/config.yaml)")
	}
	if c.Email == "" {
		return errors.New("email is required (set CONFLUENCE_EMAIL or email in ~/.confluence/config.yaml)")
	}
	if c.Token == "" {
		return errors.New("token is required (set CONFLUENCE_TOKEN or token in ~/.confluence/config.yaml)")
	}
	return nil
}

// EnsureConfigDir creates the ~/.confluence directory with 0700 permissions if
// it does not already exist. Returns an error on failure.
func EnsureConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determine home directory: %w", err)
	}

	dir := filepath.Join(home, configDirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config directory %s: %w", dir, err)
	}

	return nil
}
