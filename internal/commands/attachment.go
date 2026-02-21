package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/kristofferrisa/confluence-cli/internal/api"
	"github.com/spf13/cobra"
)

var attachmentCmd = &cobra.Command{
	Use:   "attachment",
	Short: "Manage page attachments",
	Long:  "Upload, list, and download attachments on Confluence pages.",
}

// ---------------------------------------------------------------------------
// attachment upload
// ---------------------------------------------------------------------------

var attachmentUploadCmd = &cobra.Command{
	Use:   "upload <page-id> <file>",
	Short: "Upload a file as a page attachment",
	Long:  "Attach a local file to a Confluence page.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pageID := args[0]
		filePath := args[1]

		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open file %s: %w", filePath, err)
		}
		defer f.Close()

		// Use only the base filename, not the full path, as the attachment name.
		filename := filePath
		for i := len(filePath) - 1; i >= 0; i-- {
			if filePath[i] == '/' || filePath[i] == '\\' {
				filename = filePath[i+1:]
				break
			}
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		attachment, err := client.UploadAttachment(ctx, pageID, filename, f)
		if err != nil {
			return fmt.Errorf("upload attachment: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Uploaded attachment %s (ID: %s) to page %s\n", attachment.Title, attachment.ID, pageID)
		return nil
	},
}

// ---------------------------------------------------------------------------
// attachment list
// ---------------------------------------------------------------------------

var attachmentListCmd = &cobra.Command{
	Use:   "list <page-id>",
	Short: "List attachments on a page",
	Long:  "List all files attached to a Confluence page.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		attachments, err := client.ListAttachments(ctx, args[0])
		if err != nil {
			return fmt.Errorf("list attachments: %w", err)
		}

		fmt.Println(formatter.FormatAttachments(attachments))
		return nil
	},
}

// ---------------------------------------------------------------------------
// attachment download
// ---------------------------------------------------------------------------

var attachmentDownloadOutputFlag string

var attachmentDownloadCmd = &cobra.Command{
	Use:   "download <page-id> <attachment-id>",
	Short: "Download an attachment from a page",
	Long:  "Download a file attachment from a Confluence page to the local filesystem.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pageID := args[0]
		attachmentID := args[1]

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.Token)
		ctx := context.Background()

		// List attachments to find the download link and title for this attachment ID.
		attachments, err := client.ListAttachments(ctx, pageID)
		if err != nil {
			return fmt.Errorf("list attachments: %w", err)
		}

		var downloadPath string
		var attachmentTitle string
		for _, a := range attachments {
			if a.ID == attachmentID {
				if a.Links != nil {
					downloadPath = a.Links.Download
				}
				attachmentTitle = a.Title
				break
			}
		}

		if downloadPath == "" {
			return fmt.Errorf("attachment %s not found on page %s", attachmentID, pageID)
		}

		outputPath := attachmentDownloadOutputFlag
		if outputPath == "" {
			outputPath = attachmentTitle
		}
		if outputPath == "" {
			outputPath = attachmentID
		}

		outFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("create output file %s: %w", outputPath, err)
		}
		defer outFile.Close()

		if err := client.DownloadAttachment(ctx, downloadPath, outFile); err != nil {
			// Remove partially written file on failure.
			_ = os.Remove(outputPath)
			return fmt.Errorf("download attachment: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Downloaded attachment %s to %s\n", attachmentID, outputPath)
		return nil
	},
}

func init() {
	attachmentDownloadCmd.Flags().StringVarP(&attachmentDownloadOutputFlag, "output", "o", "", "output file path (default: attachment title)")

	attachmentCmd.AddCommand(attachmentUploadCmd)
	attachmentCmd.AddCommand(attachmentListCmd)
	attachmentCmd.AddCommand(attachmentDownloadCmd)
	rootCmd.AddCommand(attachmentCmd)
}
