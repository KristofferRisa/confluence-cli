package commands

import (
	"fmt"

	"github.com/kristofferrisa/confluence-cli/internal/config"
	"github.com/kristofferrisa/confluence-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	formatFlag string
	cfg        *config.Config
	formatter  output.Formatter
)

var rootCmd = &cobra.Command{
	Use:   "cfluence",
	Short: "Confluence CLI - Manage Confluence pages with markdown",
	Long: `A command-line tool for managing Confluence Cloud pages using markdown files.

Author pages as .md files with YAML frontmatter, push to Confluence,
pull pages back as markdown, search content, manage labels and attachments.

Set credentials via environment variables:
  CONFLUENCE_BASE_URL  - Your Confluence instance URL
  CONFLUENCE_EMAIL     - Your Atlassian email
  CONFLUENCE_TOKEN     - Your API token`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading and validation for config, version, and help commands
		name := cmd.Name()
		if name == "help" || name == "version" {
			return nil
		}
		if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			return nil
		}
		if name == "config" {
			return nil
		}

		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override format if --format flag was explicitly set
		if formatFlag != "" {
			cfg.Format = formatFlag
		}

		formatter = output.New(cfg.Format)

		// Skip validation for config and version commands
		if name == "init" || name == "show" || name == "set" || name == "path" {
			return nil
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}

		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", fmt.Sprintf("config file path (default %q)", config.DefaultConfigPath()))
	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "", "output format: pretty, json, markdown")
}

