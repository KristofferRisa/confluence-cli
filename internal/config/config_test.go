package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kristofferrisa/confluence-cli/internal/config"
)

// writeConfigFile writes a YAML config file to a temporary directory and
// sets HOME to that directory so viper finds it via the default search path.
// Returns the path to the written config file.
func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	confDir := filepath.Join(tmpDir, ".confluence")
	if err := os.MkdirAll(confDir, 0700); err != nil {
		t.Fatalf("create temp config dir: %v", err)
	}

	confFile := filepath.Join(confDir, "config.yaml")
	if err := os.WriteFile(confFile, []byte(content), 0600); err != nil {
		t.Fatalf("write temp config file: %v", err)
	}

	// Point HOME at the temp dir so viper finds the file in the default location.
	t.Setenv("HOME", tmpDir)

	return confFile
}

// unsetConfluenceEnv removes all CONFLUENCE_* env vars for the duration of the
// test, restoring them in cleanup.
func unsetConfluenceEnv(t *testing.T) {
	t.Helper()

	vars := []string{
		"CONFLUENCE_BASE_URL",
		"CONFLUENCE_EMAIL",
		"CONFLUENCE_TOKEN",
		"CONFLUENCE_SPACE",
	}

	for _, v := range vars {
		prev, existed := os.LookupEnv(v)
		if err := os.Unsetenv(v); err != nil {
			t.Fatalf("unsetenv %s: %v", v, err)
		}
		if existed {
			captured := prev
			t.Cleanup(func() { _ = os.Setenv(v, captured) })
		} else {
			captured := v
			t.Cleanup(func() { _ = os.Unsetenv(captured) })
		}
	}
}

func TestLoad_EnvVarTakesPriority(t *testing.T) {
	// Write a config file with one set of values...
	writeConfigFile(t, `
base_url: https://from-file.atlassian.net
email: file@example.com
token: file-token
space: FILE
`)

	// ...then override with env vars (env must win).
	t.Setenv("CONFLUENCE_BASE_URL", "https://from-env.atlassian.net")
	t.Setenv("CONFLUENCE_EMAIL", "env@example.com")
	t.Setenv("CONFLUENCE_TOKEN", "env-token")
	t.Setenv("CONFLUENCE_SPACE", "ENV")

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.BaseURL != "https://from-env.atlassian.net" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://from-env.atlassian.net")
	}
	if cfg.Email != "env@example.com" {
		t.Errorf("Email = %q, want %q", cfg.Email, "env@example.com")
	}
	if cfg.Token != "env-token" {
		t.Errorf("Token = %q, want %q", cfg.Token, "env-token")
	}
	if cfg.Space != "ENV" {
		t.Errorf("Space = %q, want %q", cfg.Space, "ENV")
	}
}

func TestLoad_ConfigFile(t *testing.T) {
	// Ensure CONFLUENCE_* env vars are absent so the config file takes effect.
	unsetConfluenceEnv(t)

	writeConfigFile(t, `
base_url: https://mysite.atlassian.net
email: user@example.com
token: secret-api-token
space: MYSPACE
format: json
`)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.BaseURL != "https://mysite.atlassian.net" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://mysite.atlassian.net")
	}
	if cfg.Email != "user@example.com" {
		t.Errorf("Email = %q, want %q", cfg.Email, "user@example.com")
	}
	if cfg.Token != "secret-api-token" {
		t.Errorf("Token = %q, want %q", cfg.Token, "secret-api-token")
	}
	if cfg.Space != "MYSPACE" {
		t.Errorf("Space = %q, want %q", cfg.Space, "MYSPACE")
	}
	if cfg.Format != "json" {
		t.Errorf("Format = %q, want %q", cfg.Format, "json")
	}
}

func TestLoad_DefaultFormat(t *testing.T) {
	// No config file and no env vars — the "pretty" default must apply.
	unsetConfluenceEnv(t)

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Format != "pretty" {
		t.Errorf("Format = %q, want %q", cfg.Format, "pretty")
	}
}

func TestLoad_ExplicitConfigFilePath(t *testing.T) {
	// Write to a non-default location and pass the path explicitly.
	unsetConfluenceEnv(t)

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "myconf.yaml")
	content := `
base_url: https://explicit.atlassian.net
email: explicit@example.com
token: explicit-token
`
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Load(%q) error: %v", cfgPath, err)
	}

	if cfg.BaseURL != "https://explicit.atlassian.net" {
		t.Errorf("BaseURL = %q, want https://explicit.atlassian.net", cfg.BaseURL)
	}
}

func TestValidate_NoToken(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "https://mysite.atlassian.net",
		Email:   "user@example.com",
		Token:   "",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() expected error for empty token, got nil")
	}
}

func TestValidate_NoBaseURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "",
		Email:   "user@example.com",
		Token:   "secret",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() expected error for empty base_url, got nil")
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "https://mysite.atlassian.net",
		Email:   "user@example.com",
		Token:   "secret",
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}
