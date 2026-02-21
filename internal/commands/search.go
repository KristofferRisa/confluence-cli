package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/kristofferrisa/confluence-cli/internal/api"
	"github.com/kristofferrisa/confluence-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	searchSpaceFlag string
	searchLimitFlag int
)

// cqlOperators are tokens that indicate the user has written raw CQL.
var cqlOperators = []string{"=", "~", " AND ", " OR ", " IN ", " NOT ", " ORDER BY "}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search Confluence content",
	Long: `Search Confluence using a query string or raw CQL expression.

If the query does not contain CQL operators it is automatically wrapped as:
  type=page AND text~"<query>"

Examples:
  confluence search "deployment guide"
  confluence search 'space=ENG AND type=page AND text~"API"'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		cql := buildCQL(args[0], searchSpaceFlag)

		opts := &models.ListOptions{
			Limit: searchLimitFlag,
		}

		results, err := client.Search(ctx, cql, opts)
		if err != nil {
			return fmt.Errorf("search: %w", err)
		}

		fmt.Println(formatter.FormatSearchResults(results))
		return nil
	},
}

// buildCQL constructs the CQL expression to send to the API.
// If the user's query already contains CQL operators it is used as-is (with an
// optional space prefix prepended). Otherwise it is treated as a plain-text
// search term and wrapped accordingly.
func buildCQL(query, spaceKey string) string {
	isCQL := false
	upper := strings.ToUpper(query)
	for _, op := range cqlOperators {
		if strings.Contains(upper, strings.ToUpper(op)) {
			isCQL = true
			break
		}
	}

	var cql string
	if isCQL {
		cql = query
	} else {
		cql = fmt.Sprintf(`type=page AND text~"%s"`, query)
	}

	if spaceKey != "" {
		cql = fmt.Sprintf(`space="%s" AND %s`, spaceKey, cql)
	}

	return cql
}

func init() {
	searchCmd.Flags().StringVarP(&searchSpaceFlag, "space", "s", "", "limit search to this space key")
	searchCmd.Flags().IntVarP(&searchLimitFlag, "limit", "l", 25, "maximum number of results to return")
	rootCmd.AddCommand(searchCmd)
}
