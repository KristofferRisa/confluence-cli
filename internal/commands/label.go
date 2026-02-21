package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/kristofferrisa/confluence-cli/internal/api"
	"github.com/spf13/cobra"
)

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage page labels",
	Long:  "Add, list, and remove labels on Confluence pages.",
}

// ---------------------------------------------------------------------------
// label add
// ---------------------------------------------------------------------------

var labelAddCmd = &cobra.Command{
	Use:   "add <page-id> <label1> [label2 ...]",
	Short: "Add labels to a page",
	Long:  "Add one or more labels to a Confluence page.",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pageID := args[0]
		labels := args[1:]

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		if err := client.AddLabels(ctx, pageID, labels); err != nil {
			return fmt.Errorf("add labels: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Added %d label(s) to page %s\n", len(labels), pageID)
		return nil
	},
}

// ---------------------------------------------------------------------------
// label list
// ---------------------------------------------------------------------------

var labelListCmd = &cobra.Command{
	Use:   "list <page-id>",
	Short: "List labels on a page",
	Long:  "List all labels attached to a Confluence page.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		labels, err := client.GetLabels(ctx, args[0])
		if err != nil {
			return fmt.Errorf("get labels: %w", err)
		}

		fmt.Println(formatter.FormatLabels(labels))
		return nil
	},
}

// ---------------------------------------------------------------------------
// label remove
// ---------------------------------------------------------------------------

var labelRemoveCmd = &cobra.Command{
	Use:   "remove <page-id> <label>",
	Short: "Remove a label from a page",
	Long:  "Remove a single label from a Confluence page.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pageID := args[0]
		label := args[1]

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		if err := client.RemoveLabel(ctx, pageID, label); err != nil {
			return fmt.Errorf("remove label: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Removed label %q from page %s\n", label, pageID)
		return nil
	},
}

func init() {
	labelCmd.AddCommand(labelAddCmd)
	labelCmd.AddCommand(labelListCmd)
	labelCmd.AddCommand(labelRemoveCmd)
	rootCmd.AddCommand(labelCmd)
}
