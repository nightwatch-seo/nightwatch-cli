package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/nightwatch-io/nightwatch-cli/internal/client"
	"github.com/nightwatch-io/nightwatch-cli/internal/output"
	"github.com/spf13/cobra"
)

var urlsCmd = &cobra.Command{
	Use:   "urls",
	Short: "Manage tracked URLs (projects) and their settings",
	Long: `Manage tracked URLs in your Nightwatch account. URLs represent the websites
you are monitoring for SEO performance.

Available subcommands:
  list     List all tracked URLs, optionally filtered by group
  get      Get details for a specific URL by ID
  create   Add a new URL to track
  update   Update settings for a tracked URL
  delete   Remove a tracked URL

Every response is a JSON envelope: {"data": ..., "meta": {...}}`,
	Annotations: map[string]string{AnnotationRequiresAuth: "true"},
}

var urlsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked URLs",
	Long: `Retrieve all tracked URLs. Optionally filter by group ID.

Examples:
  nightwatch urls list
  nightwatch urls list --group-id 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("group-id"); v != "" {
			q.Set("group_id", v)
		}

		if isDryRun(cmd) {
			return printDryRunGet(cmd, "/urls", q)
		}

		cl, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := cl.Get(cmd.Context(), "/urls", q)
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

var urlsGetCmd = &cobra.Command{
	Use:   "get ID",
	Short: "Get details for a specific tracked URL",
	Long: `Retrieve details for a single tracked URL by its numeric ID.

Example:
  nightwatch urls get 456`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		endpoint := fmt.Sprintf("/urls/%s", id)

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

		output.PrintSingle(cmd.OutOrStdout(), data)
		return nil
	},
}

var urlsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add a new URL to track",
	Long: `Create a new tracked URL with the specified settings.

Required flags:
  --url              The URL to track (e.g. "https://example.com")
  --group-id         ID of the URL group to add this URL to
  --country-code     Two-letter country code (e.g. "us", "gb", "de")
  --language-code    Two-letter language code (e.g. "en", "de", "fr")

Optional flags:
  --custom-name           Custom display name for the URL
  --places-match          Google Places match string
  --match-nested-urls     Track nested URLs under this domain
  --match-subdomains      Track subdomains of this domain

Examples:
  nightwatch urls create --url "https://example.com" --group-id 42 --country-code us --language-code en
  nightwatch urls create --url "https://shop.example.com" --group-id 42 --country-code gb --language-code en --custom-name "UK Shop"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlFlag, _ := cmd.Flags().GetString("url")
		groupID, _ := cmd.Flags().GetString("group-id")
		countryCode, _ := cmd.Flags().GetString("country-code")
		languageCode, _ := cmd.Flags().GetString("language-code")

		if urlFlag == "" {
			output.PrintErrorTyped(os.Stderr, "--url is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}
		if groupID == "" {
			output.PrintErrorTyped(os.Stderr, "--group-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}
		if countryCode == "" {
			output.PrintErrorTyped(os.Stderr, "--country-code is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}
		if languageCode == "" {
			output.PrintErrorTyped(os.Stderr, "--language-code is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		urlData := map[string]any{
			"url":            urlFlag,
			"url_group_id":   groupID,
			"country_code":   countryCode,
			"language_code":  languageCode,
		}

		if v, _ := cmd.Flags().GetString("custom-name"); v != "" {
			urlData["custom_name"] = v
		}
		if v, _ := cmd.Flags().GetString("places-match"); v != "" {
			urlData["places_match"] = v
		}
		if cmd.Flags().Changed("match-nested-urls") {
			v, _ := cmd.Flags().GetBool("match-nested-urls")
			urlData["match_nested_urls"] = v
		}
		if cmd.Flags().Changed("match-subdomains") {
			v, _ := cmd.Flags().GetBool("match-subdomains")
			urlData["match_subdomains"] = v
		}

		body := map[string]any{"url": urlData}

		if isDryRun(cmd) {
			return printDryRunPost(cmd, "/urls", body)
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

		resp, err := cl.Post(cmd.Context(), "/urls", jsonBody)
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

var urlsUpdateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update settings for a tracked URL",
	Long: `Update the settings of an existing tracked URL by its numeric ID.
At least one flag must be provided.

Optional flags:
  --url                          New URL value
  --custom-name                  Custom display name
  --country-code                 Two-letter country code
  --language-code                Two-letter language code
  --places-match                 Google Places match string
  --match-nested-urls            Track nested URLs
  --match-subdomains             Track subdomains
  --group-id                     Move to a different URL group
  --include-local-pack           Include local pack in main position
  --include-places-image         Include places image in main position
  --include-featured-snippet     Include featured snippet in main position
  --include-knowledge-panel      Include knowledge panel in main position
  --site-audit-interval          Site audit interval

Example:
  nightwatch urls update 456 --custom-name "Updated Name" --country-code gb`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		endpoint := fmt.Sprintf("/urls/%s", id)

		urlData := map[string]any{}

		if v, _ := cmd.Flags().GetString("url"); cmd.Flags().Changed("url") {
			urlData["url"] = v
		}
		if v, _ := cmd.Flags().GetString("custom-name"); cmd.Flags().Changed("custom-name") {
			urlData["custom_name"] = v
		}
		if v, _ := cmd.Flags().GetString("country-code"); cmd.Flags().Changed("country-code") {
			urlData["country_code"] = v
		}
		if v, _ := cmd.Flags().GetString("language-code"); cmd.Flags().Changed("language-code") {
			urlData["language_code"] = v
		}
		if v, _ := cmd.Flags().GetString("places-match"); cmd.Flags().Changed("places-match") {
			urlData["places_match"] = v
		}
		if cmd.Flags().Changed("match-nested-urls") {
			v, _ := cmd.Flags().GetBool("match-nested-urls")
			urlData["match_nested_urls"] = v
		}
		if cmd.Flags().Changed("match-subdomains") {
			v, _ := cmd.Flags().GetBool("match-subdomains")
			urlData["match_subdomains"] = v
		}
		if v, _ := cmd.Flags().GetString("group-id"); cmd.Flags().Changed("group-id") {
			urlData["url_group_id"] = v
		}
		if cmd.Flags().Changed("include-local-pack") {
			v, _ := cmd.Flags().GetBool("include-local-pack")
			urlData["include_local_pack_in_main_position"] = v
		}
		if cmd.Flags().Changed("include-places-image") {
			v, _ := cmd.Flags().GetBool("include-places-image")
			urlData["include_places_image_in_main_position"] = v
		}
		if cmd.Flags().Changed("include-featured-snippet") {
			v, _ := cmd.Flags().GetBool("include-featured-snippet")
			urlData["include_featured_snippet_in_main_position"] = v
		}
		if cmd.Flags().Changed("include-knowledge-panel") {
			v, _ := cmd.Flags().GetBool("include-knowledge-panel")
			urlData["include_knowledge_panel_in_main_position"] = v
		}
		if v, _ := cmd.Flags().GetString("site-audit-interval"); cmd.Flags().Changed("site-audit-interval") {
			urlData["site_audit_interval"] = v
		}

		if len(urlData) == 0 {
			output.PrintErrorTyped(os.Stderr, "at least one field to update is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		body := map[string]any{"url": urlData}

		if isDryRun(cmd) {
			return printDryRunPut(cmd, endpoint, body)
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

		resp, err := cl.Put(cmd.Context(), endpoint, jsonBody)
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

var urlsDeleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "Delete a tracked URL",
	Long: `Delete a tracked URL by its numeric ID. This removes all associated
keywords, views, and historical data.

Example:
  nightwatch urls delete 456`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		endpoint := fmt.Sprintf("/urls/%s", id)

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
			"status": "deleted",
			"id":     id,
		})
		return nil
	},
}

func init() {
	// list flags
	urlsListCmd.Flags().String("group-id", "", "Filter URLs by group ID")

	// create flags
	urlsCreateCmd.Flags().String("url", "", "URL to track (required)")
	urlsCreateCmd.Flags().String("group-id", "", "URL group ID (required)")
	urlsCreateCmd.Flags().String("country-code", "", "Two-letter country code, e.g. us, gb (required)")
	urlsCreateCmd.Flags().String("language-code", "", "Two-letter language code, e.g. en, de (required)")
	urlsCreateCmd.Flags().String("custom-name", "", "Custom display name for the URL")
	urlsCreateCmd.Flags().String("places-match", "", "Google Places match string")
	urlsCreateCmd.Flags().Bool("match-nested-urls", false, "Track nested URLs under this domain")
	urlsCreateCmd.Flags().Bool("match-subdomains", false, "Track subdomains of this domain")

	// update flags
	urlsUpdateCmd.Flags().String("url", "", "New URL value")
	urlsUpdateCmd.Flags().String("custom-name", "", "Custom display name")
	urlsUpdateCmd.Flags().String("country-code", "", "Two-letter country code")
	urlsUpdateCmd.Flags().String("language-code", "", "Two-letter language code")
	urlsUpdateCmd.Flags().String("places-match", "", "Google Places match string")
	urlsUpdateCmd.Flags().Bool("match-nested-urls", false, "Track nested URLs")
	urlsUpdateCmd.Flags().Bool("match-subdomains", false, "Track subdomains")
	urlsUpdateCmd.Flags().String("group-id", "", "Move to a different URL group")
	urlsUpdateCmd.Flags().Bool("include-local-pack", false, "Include local pack in main position")
	urlsUpdateCmd.Flags().Bool("include-places-image", false, "Include places image in main position")
	urlsUpdateCmd.Flags().Bool("include-featured-snippet", false, "Include featured snippet in main position")
	urlsUpdateCmd.Flags().Bool("include-knowledge-panel", false, "Include knowledge panel in main position")
	urlsUpdateCmd.Flags().String("site-audit-interval", "", "Site audit interval")

	urlsCmd.AddCommand(urlsListCmd)
	urlsCmd.AddCommand(urlsGetCmd)
	urlsCmd.AddCommand(urlsCreateCmd)
	urlsCmd.AddCommand(urlsUpdateCmd)
	urlsCmd.AddCommand(urlsDeleteCmd)
	rootCmd.AddCommand(urlsCmd)
}
