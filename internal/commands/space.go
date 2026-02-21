package commands

import (
	"context"
	"fmt"

	"github.com/kristofferrisa/confluence-cli/internal/api"
	"github.com/kristofferrisa/confluence-cli/internal/models"
	"github.com/spf13/cobra"
)

var spaceCmd = &cobra.Command{
	Use:   "space",
	Short: "Manage Confluence spaces",
	Long:  "List and inspect Confluence spaces.",
}

// ---------------------------------------------------------------------------
// space list
// ---------------------------------------------------------------------------

var spaceListLimitFlag int

var spaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Confluence spaces",
	Long:  "List all accessible Confluence spaces.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		opts := &models.ListOptions{
			Limit: spaceListLimitFlag,
		}

		spaceList, err := client.ListSpaces(ctx, opts)
		if err != nil {
			return fmt.Errorf("list spaces: %w", err)
		}

		fmt.Println(formatter.FormatSpaces(spaceList.Results))
		return nil
	},
}

// ---------------------------------------------------------------------------
// space get
// ---------------------------------------------------------------------------

var spaceGetCmd = &cobra.Command{
	Use:   "get <space-key>",
	Short: "Get a Confluence space by key",
	Long:  "Retrieve a space by its key and display its details.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		space, err := client.GetSpaceByKey(ctx, args[0])
		if err != nil {
			return fmt.Errorf("get space %q: %w", args[0], err)
		}

		fmt.Println(formatter.FormatSpace(space))
		return nil
	},
}

func init() {
	spaceListCmd.Flags().IntVarP(&spaceListLimitFlag, "limit", "l", 25, "maximum number of spaces to return")

	spaceCmd.AddCommand(spaceListCmd)
	spaceCmd.AddCommand(spaceGetCmd)
	rootCmd.AddCommand(spaceCmd)
}
