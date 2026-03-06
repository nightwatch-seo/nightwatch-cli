package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nightwatch-io/nightwatch-cli/internal/client"
	"github.com/nightwatch-io/nightwatch-cli/internal/output"
	"github.com/spf13/cobra"
)

var competitorsCmd = &cobra.Command{
	Use:   "competitors",
	Short: "Track and manage competitor URLs",
	Long: `Manage competitor URLs for a tracked URL in your Nightwatch account.
Competitors let you compare your ranking performance against other sites.

Available subcommands:
  list     List competitors for a URL
  add      Add a competitor to a URL
  remove   Remove a competitor from a URL

Every response is a JSON envelope: {"data": ..., "meta": {...}}`,
	Annotations: map[string]string{AnnotationRequiresAuth: "true"},
}

var competitorsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List competitors for a URL",
	Long: `Retrieve all competitors tracked for a specific URL.

Required flags:
  --url-id    ID of the tracked URL

Example:
  nightwatch competitors list --url-id 123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlID, _ := cmd.Flags().GetString("url-id")
		if urlID == "" {
			output.PrintErrorTyped(os.Stderr, "--url-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		endpoint := fmt.Sprintf("/urls/%s/competitors", urlID)

		if isDryRun(cmd) {
			return printDryRunGet(cmd, endpoint, nil)
		}

		cl, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := cl.Get(cmd.Context(), endpoint, nil)
		if err != nil {
			return handleAPIError(err)
		}

		data, err := parseResponseData(resp.Body)
		if err != nil {
			output.PrintErrorTyped(os.Stderr, err.Error(), 1, client.ErrorTypeClient)
			return &exitError{code: 1}
		}

		output.PrintList(cmd.OutOrStdout(), data)
		return nil
	},
}

var competitorsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a competitor to a URL",
	Long: `Add a competitor URL to track against a monitored URL.

Required flags:
  --url-id    ID of the tracked URL
  --url       Competitor URL to add

Optional flags:
  --custom-name          Custom display name
  --match-subdomains     Match subdomains
  --match-nested-urls    Match nested URLs
  --places-match         Google Places match string
  --places-image-title   Google Places image title

Examples:
  nightwatch competitors add --url-id 123 --url "https://competitor.com"
  nightwatch competitors add --url-id 123 --url "https://competitor.com" --custom-name "Main Competitor" --match-subdomains`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlID, _ := cmd.Flags().GetString("url-id")
		if urlID == "" {
			output.PrintErrorTyped(os.Stderr, "--url-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		compURL, _ := cmd.Flags().GetString("url")
		if compURL == "" {
			output.PrintErrorTyped(os.Stderr, "--url is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		compData := map[string]any{
			"url": compURL,
		}
		if v, _ := cmd.Flags().GetString("custom-name"); v != "" {
			compData["custom_name"] = v
		}
		if cmd.Flags().Changed("match-subdomains") {
			v, _ := cmd.Flags().GetBool("match-subdomains")
			compData["match_subdomains"] = v
		}
		if cmd.Flags().Changed("match-nested-urls") {
			v, _ := cmd.Flags().GetBool("match-nested-urls")
			compData["match_nested_urls"] = v
		}
		if v, _ := cmd.Flags().GetString("places-match"); v != "" {
			compData["places_match"] = v
		}
		if v, _ := cmd.Flags().GetString("places-image-title"); v != "" {
			compData["places_image_title"] = v
		}

		endpoint := fmt.Sprintf("/urls/%s/competitors", urlID)
		body := map[string]any{"competitor": compData}

		if isDryRun(cmd) {
			return printDryRunPost(cmd, endpoint, body)
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			output.PrintErrorTyped(os.Stderr, err.Error(), 1, client.ErrorTypeClient)
			return &exitError{code: 1}
		}

		cl, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := cl.Post(cmd.Context(), endpoint, jsonBody)
		if err != nil {
			return handleAPIError(err)
		}

		data, err := parseResponseData(resp.Body)
		if err != nil {
			output.PrintErrorTyped(os.Stderr, err.Error(), 1, client.ErrorTypeClient)
			return &exitError{code: 1}
		}

		output.PrintSingle(cmd.OutOrStdout(), data)
		return nil
	},
}

var competitorsRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a competitor from a URL",
	Long: `Remove a competitor from a tracked URL by competitor ID.

Required flags:
  --url-id          ID of the tracked URL
  --competitor-id   ID of the competitor to remove

Example:
  nightwatch competitors remove --url-id 123 --competitor-id 456`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlID, _ := cmd.Flags().GetString("url-id")
		if urlID == "" {
			output.PrintErrorTyped(os.Stderr, "--url-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		compID, _ := cmd.Flags().GetString("competitor-id")
		if compID == "" {
			output.PrintErrorTyped(os.Stderr, "--competitor-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		endpoint := fmt.Sprintf("/urls/%s/competitors/%s", urlID, compID)

		if isDryRun(cmd) {
			return printDryRunDelete(cmd, endpoint)
		}

		cl, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}

		_, err = cl.Delete(cmd.Context(), endpoint)
		if err != nil {
			return handleAPIError(err)
		}

		output.PrintData(cmd.OutOrStdout(), map[string]string{
			"status":        "deleted",
			"url_id":        urlID,
			"competitor_id": compID,
		})
		return nil
	},
}

func init() {
	// list flags
	competitorsListCmd.Flags().String("url-id", "", "ID of the tracked URL (required)")

	// add flags
	competitorsAddCmd.Flags().String("url-id", "", "ID of the tracked URL (required)")
	competitorsAddCmd.Flags().String("url", "", "Competitor URL to add (required)")
	competitorsAddCmd.Flags().String("custom-name", "", "Custom display name")
	competitorsAddCmd.Flags().Bool("match-subdomains", false, "Match subdomains")
	competitorsAddCmd.Flags().Bool("match-nested-urls", false, "Match nested URLs")
	competitorsAddCmd.Flags().String("places-match", "", "Google Places match string")
	competitorsAddCmd.Flags().String("places-image-title", "", "Google Places image title")

	// remove flags
	competitorsRemoveCmd.Flags().String("url-id", "", "ID of the tracked URL (required)")
	competitorsRemoveCmd.Flags().String("competitor-id", "", "ID of the competitor to remove (required)")

	competitorsCmd.AddCommand(competitorsListCmd)
	competitorsCmd.AddCommand(competitorsAddCmd)
	competitorsCmd.AddCommand(competitorsRemoveCmd)
	rootCmd.AddCommand(competitorsCmd)
}
