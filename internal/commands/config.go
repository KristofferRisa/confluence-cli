package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kristofferrisa/confluence-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage confluence-cli configuration",
	Long:  "Initialize, view, and update the confluence-cli configuration file.",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup wizard",
	Long:  "Walk through an interactive setup to create the configuration file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.EnsureConfigDir(); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		scanner := bufio.NewScanner(os.Stdin)

		fmt.Fprintln(os.Stderr, "Confluence CLI setup wizard")
		fmt.Fprintln(os.Stderr, "Press Enter to keep the current value shown in brackets.")
		fmt.Fprintln(os.Stderr)

		baseURL := prompt(scanner, "Confluence base URL (e.g. https://your-org.atlassian.net)", "")
		email := prompt(scanner, "Atlassian email address", "")
		token := prompt(scanner, "Atlassian API token", "")
		space := prompt(scanner, "Default space key (optional)", "")
		format := prompt(scanner, "Output format [pretty/json/markdown]", "pretty")

		if format == "" {
			format = "pretty"
		}

		v := viper.New()
		v.Set("base_url", baseURL)
		v.Set("email", email)
		v.Set("token", token)
		v.Set("space", space)
		v.Set("format", format)

		configPath := config.DefaultConfigPath()
		v.SetConfigFile(configPath)
		v.SetConfigType("yaml")

		if err := v.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Fprintf(os.Stderr, "\nConfig written to %s\n", configPath)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  "Show all configuration values. The API token is masked for security.",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		maskedToken := maskToken(c.Token)

		fmt.Printf("base_url : %s\n", c.BaseURL)
		fmt.Printf("email    : %s\n", c.Email)
		fmt.Printf("token    : %s\n", maskedToken)
		fmt.Printf("space    : %s\n", c.Space)
		fmt.Printf("format   : %s\n", c.Format)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a single configuration value by key.

Supported keys: base_url, email, token, space, format`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		validKeys := map[string]bool{
			"base_url": true,
			"email":    true,
			"token":    true,
			"space":    true,
			"format":   true,
		}

		if !validKeys[key] {
			return fmt.Errorf("unsupported config key %q, supported keys: base_url, email, token, space, format", key)
		}

		if err := config.EnsureConfigDir(); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		configPath := cfgFile
		if configPath == "" {
			configPath = config.DefaultConfigPath()
		}

		v := viper.New()
		v.SetConfigFile(configPath)
		v.SetConfigType("yaml")

		// Load existing config if it exists; ignore error if file doesn't exist yet
		_ = v.ReadInConfig()

		v.Set(key, value)

		if err := v.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Set %s in %s\n", key, configPath)
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the config file path",
	Long:  "Print the path to the configuration file being used.",
	Run: func(cmd *cobra.Command, args []string) {
		p := cfgFile
		if p == "" {
			p = config.DefaultConfigPath()
		}
		fmt.Println(p)
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}

// prompt reads a line from stdin after printing a prompt. If the user enters
// nothing, the provided default value is returned.
func prompt(scanner *bufio.Scanner, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Fprintf(os.Stderr, "%s [%s]: ", label, defaultVal)
	} else {
		fmt.Fprintf(os.Stderr, "%s: ", label)
	}
	scanner.Scan()
	line := strings.TrimSpace(scanner.Text())
	if line == "" {
		return defaultVal
	}
	return line
}

// maskToken returns the token with all but the last 4 characters replaced by
// asterisks. Returns "****" for tokens shorter than 4 characters.
func maskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(token)-4) + token[len(token)-4:]
}
