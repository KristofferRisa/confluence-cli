package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/kristofferrisa/confluence-cli/internal/api"
	"github.com/kristofferrisa/confluence-cli/internal/converter"
	"github.com/kristofferrisa/confluence-cli/internal/models"
	"github.com/spf13/cobra"
)

var pageCmd = &cobra.Command{
	Use:   "page",
	Short: "Manage Confluence pages",
	Long:  "Create, update, retrieve, list, delete, and display page trees.",
}

// ---------------------------------------------------------------------------
// page push
// ---------------------------------------------------------------------------

var pagePushCmd = &cobra.Command{
	Use:   "push <file.md> [file2.md ...]",
	Short: "Push one or more markdown files to Confluence",
	Long: `Push markdown files to Confluence. If the file's frontmatter contains a
page_id the page is updated; otherwise a new page is created and page_id is
written back to the file.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		for _, filePath := range args {
			if err := pushFile(ctx, client, filePath); err != nil {
				return fmt.Errorf("push %s: %w", filePath, err)
			}
		}
		return nil
	},
}

func pushFile(ctx context.Context, client *api.Client, filePath string) error {
	fm, body, err := converter.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parse file: %w", err)
	}

	storageBody, err := converter.MarkdownToStorage(body)
	if err != nil {
		return fmt.Errorf("convert markdown: %w", err)
	}

	var page *models.Page

	if fm.PageID != "" {
		// UPDATE existing page
		existing, err := client.GetPage(ctx, fm.PageID)
		if err != nil {
			return fmt.Errorf("get page %s: %w", fm.PageID, err)
		}

		currentVersion := 1
		if existing.Version != nil {
			currentVersion = existing.Version.Number
		}

		req := &models.UpdatePageRequest{
			ID:     fm.PageID,
			Status: "current",
			Title:  fm.Title,
			Body: models.CreatePageBody{
				Representation: "storage",
				Value:          storageBody,
			},
			Version: models.UpdateVersion{
				Number: currentVersion + 1,
			},
		}

		page, err = client.UpdatePage(ctx, fm.PageID, req)
		if err != nil {
			return fmt.Errorf("update page: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Updated page %s: %s\n", page.ID, pageWebURL(page, cfg.BaseURL))
	} else {
		// CREATE new page
		spaceKey := fm.Space
		if spaceKey == "" {
			spaceKey = cfg.Space
		}
		if spaceKey == "" {
			return fmt.Errorf("space is required: set frontmatter 'space' or configure a default space")
		}

		space, err := client.GetSpaceByKey(ctx, spaceKey)
		if err != nil {
			return fmt.Errorf("get space %q: %w", spaceKey, err)
		}

		req := &models.CreatePageRequest{
			SpaceID:  space.ID,
			Status:   "current",
			Title:    fm.Title,
			ParentID: fm.ParentID,
			Body: models.CreatePageBody{
				Representation: "storage",
				Value:          storageBody,
			},
		}

		page, err = client.CreatePage(ctx, req)
		if err != nil {
			return fmt.Errorf("create page: %w", err)
		}

		// Write page_id back to the file
		fm.PageID = page.ID
		if err := converter.WriteFile(filePath, fm, body); err != nil {
			return fmt.Errorf("write page_id back to file: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Created page %s: %s\n", page.ID, pageWebURL(page, cfg.BaseURL))
	}

	// Apply labels if present
	if len(fm.Labels) > 0 {
		if err := client.AddLabels(ctx, page.ID, fm.Labels); err != nil {
			return fmt.Errorf("add labels: %w", err)
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// page pull
// ---------------------------------------------------------------------------

var pagePullOutputFlag string

var pagePullCmd = &cobra.Command{
	Use:   "pull <page-id>",
	Short: "Pull a Confluence page as a markdown file",
	Long:  "Download a Confluence page and write it as a markdown file with YAML frontmatter.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()
		pageID := args[0]

		page, err := client.GetPage(ctx, pageID)
		if err != nil {
			return fmt.Errorf("get page: %w", err)
		}

		// Convert storage format to markdown
		var markdownBody string
		if page.Body != nil && page.Body.Storage != nil {
			markdownBody, err = converter.StorageToMarkdown(page.Body.Storage.Value)
			if err != nil {
				return fmt.Errorf("convert page body: %w", err)
			}
		}

		// Resolve space key from space ID
		spaceKey := ""
		if page.SpaceID != "" {
			space, err := client.GetSpace(ctx, page.SpaceID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not resolve space key for space ID %s: %v\n", page.SpaceID, err)
			} else {
				spaceKey = space.Key
			}
		}

		// Fetch labels
		labels, err := client.GetLabels(ctx, pageID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not fetch labels: %v\n", err)
			labels = nil
		}

		labelNames := make([]string, 0, len(labels))
		for _, l := range labels {
			labelNames = append(labelNames, l.Name)
		}

		fm := &converter.Frontmatter{
			Title:    page.Title,
			Space:    spaceKey,
			PageID:   page.ID,
			ParentID: page.ParentID,
			Labels:   labelNames,
		}

		outputPath := pagePullOutputFlag
		if outputPath == "" {
			outputPath = pageID + ".md"
		}

		if err := converter.WriteFile(outputPath, fm, markdownBody); err != nil {
			return fmt.Errorf("write file: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Pulled page %s to %s\n", pageID, outputPath)
		return nil
	},
}

// ---------------------------------------------------------------------------
// page get
// ---------------------------------------------------------------------------

var pageGetCmd = &cobra.Command{
	Use:   "get <page-id>",
	Short: "Get a single Confluence page",
	Long:  "Retrieve a page by ID and print it using the configured output format.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		page, err := client.GetPage(ctx, args[0])
		if err != nil {
			return fmt.Errorf("get page: %w", err)
		}

		fmt.Println(formatter.FormatPage(page))
		return nil
	},
}

// ---------------------------------------------------------------------------
// page list
// ---------------------------------------------------------------------------

var (
	pageListSpaceFlag  string
	pageListLimitFlag  int
	pageListStatusFlag string
)

var pageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pages in a space",
	Long:  "List all pages in a Confluence space.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		spaceKey := pageListSpaceFlag
		if spaceKey == "" {
			spaceKey = cfg.Space
		}
		if spaceKey == "" {
			return fmt.Errorf("space is required: use --space flag or configure a default space")
		}

		space, err := client.GetSpaceByKey(ctx, spaceKey)
		if err != nil {
			return fmt.Errorf("get space %q: %w", spaceKey, err)
		}

		opts := &models.ListOptions{
			Limit:  pageListLimitFlag,
			Status: pageListStatusFlag,
		}

		pageList, err := client.ListPages(ctx, space.ID, opts)
		if err != nil {
			return fmt.Errorf("list pages: %w", err)
		}

		fmt.Println(formatter.FormatPages(pageList.Results))
		return nil
	},
}

// ---------------------------------------------------------------------------
// page delete
// ---------------------------------------------------------------------------

var pageDeleteForceFlag bool

var pageDeleteCmd = &cobra.Command{
	Use:   "delete <page-id>",
	Short: "Delete a Confluence page",
	Long:  "Delete a page by ID. Prompts for confirmation unless --force is set.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pageID := args[0]

		if !pageDeleteForceFlag {
			fmt.Fprintf(os.Stderr, "Delete page %s? [y/N]: ", pageID)
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				fmt.Fprintln(os.Stderr, "Aborted.")
				return nil
			}
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		if err := client.DeletePage(ctx, pageID); err != nil {
			return fmt.Errorf("delete page: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Deleted page %s\n", pageID)
		return nil
	},
}

// ---------------------------------------------------------------------------
// page tree
// ---------------------------------------------------------------------------

var (
	pageTreeSpaceFlag string
	pageTreeDepthFlag int
)

var pageTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display the page hierarchy of a space",
	Long:  "Render a tree view of the page hierarchy for a Confluence space.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		spaceKey := pageTreeSpaceFlag
		if spaceKey == "" {
			spaceKey = cfg.Space
		}
		if spaceKey == "" {
			return fmt.Errorf("space is required: use --space flag or configure a default space")
		}

		space, err := client.GetSpaceByKey(ctx, spaceKey)
		if err != nil {
			return fmt.Errorf("get space %q: %w", spaceKey, err)
		}

		if space.HomepageID == "" {
			return fmt.Errorf("space %q has no homepage; cannot build tree", spaceKey)
		}

		tree, err := buildTree(ctx, client, space.HomepageID, 0, pageTreeDepthFlag)
		if err != nil {
			return fmt.Errorf("build tree: %w", err)
		}

		fmt.Println(formatter.FormatPageTree(tree, cfg.BaseURL))
		return nil
	},
}

// buildTree recursively fetches page children up to maxDepth levels deep.
func buildTree(ctx context.Context, client *api.Client, pageID string, depth, maxDepth int) (*models.PageTree, error) {
	page, err := client.GetPage(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("get page %s: %w", pageID, err)
	}

	tree := &models.PageTree{
		Page: *page,
	}

	if depth >= maxDepth {
		return tree, nil
	}

	childList, err := client.GetPageChildren(ctx, pageID, &models.ListOptions{Limit: 250})
	if err != nil {
		return nil, fmt.Errorf("get children of %s: %w", pageID, err)
	}

	for _, child := range childList.Results {
		childTree, err := buildTree(ctx, client, child.ID, depth+1, maxDepth)
		if err != nil {
			return nil, err
		}
		tree.Children = append(tree.Children, *childTree)
	}

	return tree, nil
}

// pageWebURL builds the web URL for a page using its Links field or falls back
// to constructing one from the base URL.
func pageWebURL(page *models.Page, baseURL string) string {
	if page.Links != nil && page.Links.WebUI != "" {
		return baseURL + page.Links.WebUI
	}
	return fmt.Sprintf("%s/wiki/pages/%s", baseURL, page.ID)
}

func init() {
	// pull flags
	pagePullCmd.Flags().StringVarP(&pagePullOutputFlag, "output", "o", "", "output file path (default: <page-id>.md)")

	// list flags
	pageListCmd.Flags().StringVarP(&pageListSpaceFlag, "space", "s", "", "space key (overrides config)")
	pageListCmd.Flags().IntVarP(&pageListLimitFlag, "limit", "l", 25, "maximum number of pages to return")
	pageListCmd.Flags().StringVar(&pageListStatusFlag, "status", "current", "page status filter")

	// delete flags
	pageDeleteCmd.Flags().BoolVar(&pageDeleteForceFlag, "force", false, "skip confirmation prompt")

	// tree flags
	pageTreeCmd.Flags().StringVarP(&pageTreeSpaceFlag, "space", "s", "", "space key (overrides config)")
	pageTreeCmd.Flags().IntVarP(&pageTreeDepthFlag, "depth", "d", 3, "maximum depth of the tree")

	pageCmd.AddCommand(pagePushCmd)
	pageCmd.AddCommand(pagePullCmd)
	pageCmd.AddCommand(pageGetCmd)
	pageCmd.AddCommand(pageListCmd)
	pageCmd.AddCommand(pageDeleteCmd)
	pageCmd.AddCommand(pageTreeCmd)
	rootCmd.AddCommand(pageCmd)
}
