package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/nightwatch-io/nightwatch-cli/internal/client"
	"github.com/nightwatch-io/nightwatch-cli/internal/output"
	"github.com/spf13/cobra"
)

var keywordsCmd = &cobra.Command{
	Use:   "keywords",
	Short: "Add, remove, update, and list tracked keywords for a URL",
	Long: `Manage keywords tracked for a specific URL in your Nightwatch account.

Available subcommands:
  list     List keywords for a URL with optional filtering, sorting, and pagination
  add      Batch-add keywords to a URL
  remove   Batch-remove keywords from a URL by keyword IDs
  update   Update a keyword's tags

Every response is a JSON envelope: {"data": ..., "meta": {...}}`,
	Annotations: map[string]string{AnnotationRequiresAuth: "true"},
}

var keywordsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List keywords for a URL",
	Long: `Retrieve keywords for a tracked URL with optional filtering, sorting,
and pagination. The response includes keyword data along with pagination info.

Required flags:
  --url-id     ID of the tracked URL

Optional flags:
  --view-id      Filter by dynamic view ID
  --page         Page number (default: 1)
  --limit        Results per page (default: 20)
  --search       Search keywords by query string
  --sort         Sort field (e.g. "position", "query")
  --direction    Sort direction: "asc" or "desc"

Examples:
  nightwatch keywords list --url-id 123
  nightwatch keywords list --url-id 123 --page 2 --limit 50
  nightwatch keywords list --url-id 123 --search "seo" --sort position --direction asc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlID, _ := cmd.Flags().GetString("url-id")
		if urlID == "" {
			output.PrintErrorTyped(os.Stderr, "--url-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		endpoint := fmt.Sprintf("/urls/%s/keywords", urlID)
		q := url.Values{}

		if v, _ := cmd.Flags().GetString("view-id"); v != "" {
			q.Set("dynamic_view_id", v)
		}
		if v, _ := cmd.Flags().GetString("page"); v != "" {
			q.Set("page", v)
		}
		if v, _ := cmd.Flags().GetString("limit"); v != "" {
			q.Set("limit", v)
		}
		if v, _ := cmd.Flags().GetString("search"); v != "" {
			q.Set("search", v)
		}
		if v, _ := cmd.Flags().GetString("sort"); v != "" {
			q.Set("sort", v)
		}
		if v, _ := cmd.Flags().GetString("direction"); v != "" {
			q.Set("direction", v)
		}

		if isDryRun(cmd) {
			return printDryRunGet(cmd, endpoint, q)
		}

		cl, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := cl.Get(cmd.Context(), endpoint, q)
		if err != nil {
			return handleAPIError(err)
		}

		data, err := parseResponseData(resp.Body)
		if err != nil {
			output.PrintErrorTyped(os.Stderr, err.Error(), 1, client.ErrorTypeClient)
			return &exitError{code: 1}
		}

		// Keywords list returns an object with items + pagination info, pass through as-is.
		output.PrintSingle(cmd.OutOrStdout(), data)
		return nil
	},
}

var keywordsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Batch-add keywords to a URL",
	Long: `Add one or more keywords to a tracked URL. Keywords are provided as a
comma-separated list and will be submitted as a batch.

Required flags:
  --url-id          ID of the tracked URL
  --keywords        Comma-separated list of keywords to add
  --google-gl       Google geolocation code (e.g. "us", "gb")
  --google-hl       Google language code (e.g. "en", "de")
  --search-engine   Search engine to track (e.g. "google", "bing", "youtube")

Optional flags:
  --mobile                Track mobile rankings
  --desktop               Track desktop rankings
  --tags                  Comma-separated tags to assign
  --adwords-location-id   Google Ads location ID for local results

Examples:
  nightwatch keywords add --url-id 123 --keywords "seo tools,keyword tracking" --google-gl us --google-hl en --search-engine google
  nightwatch keywords add --url-id 123 --keywords "local seo" --google-gl us --google-hl en --search-engine google --mobile --tags "local,priority"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlID, _ := cmd.Flags().GetString("url-id")
		if urlID == "" {
			output.PrintErrorTyped(os.Stderr, "--url-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		keywords, _ := cmd.Flags().GetString("keywords")
		if keywords == "" {
			output.PrintErrorTyped(os.Stderr, "--keywords is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		googleGL, _ := cmd.Flags().GetString("google-gl")
		if googleGL == "" {
			output.PrintErrorTyped(os.Stderr, "--google-gl is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		googleHL, _ := cmd.Flags().GetString("google-hl")
		if googleHL == "" {
			output.PrintErrorTyped(os.Stderr, "--google-hl is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		searchEngine, _ := cmd.Flags().GetString("search-engine")
		if searchEngine == "" {
			output.PrintErrorTyped(os.Stderr, "--search-engine is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		// Convert comma-separated keywords to newline-separated (API format).
		keywordList := strings.Split(keywords, ",")
		for i := range keywordList {
			keywordList[i] = strings.TrimSpace(keywordList[i])
		}
		keywordsNewline := strings.Join(keywordList, "\n")

		body := map[string]any{
			"keywords":      keywordsNewline,
			"google_gl":     googleGL,
			"google_hl":     googleHL,
			"search_engine": searchEngine,
		}

		if cmd.Flags().Changed("mobile") {
			v, _ := cmd.Flags().GetBool("mobile")
			body["mobile"] = v
		}
		if cmd.Flags().Changed("desktop") {
			v, _ := cmd.Flags().GetBool("desktop")
			body["desktop"] = v
		}
		if v, _ := cmd.Flags().GetString("tags"); v != "" {
			tags := strings.Split(v, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			body["tags"] = tags
		}
		if v, _ := cmd.Flags().GetString("adwords-location-id"); v != "" {
			body["adwords_location_id"] = v
		}

		endpoint := fmt.Sprintf("/urls/%s/keywords/batch_create", urlID)

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

var keywordsRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Batch-remove keywords from a URL",
	Long: `Remove one or more keywords from a tracked URL by their keyword IDs.

Required flags:
  --url-id         ID of the tracked URL
  --keyword-ids    Comma-separated list of keyword IDs to remove

Example:
  nightwatch keywords remove --url-id 123 --keyword-ids 456,789,101`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlID, _ := cmd.Flags().GetString("url-id")
		if urlID == "" {
			output.PrintErrorTyped(os.Stderr, "--url-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		keywordIDs, _ := cmd.Flags().GetString("keyword-ids")
		if keywordIDs == "" {
			output.PrintErrorTyped(os.Stderr, "--keyword-ids is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		ids := strings.Split(keywordIDs, ",")
		for i := range ids {
			ids[i] = strings.TrimSpace(ids[i])
		}

		body := map[string]any{
			"keyword_ids": ids,
		}

		endpoint := fmt.Sprintf("/urls/%s/keywords/batch_destroy", urlID)

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

var keywordsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a keyword's tags",
	Long: `Update the tags for a specific keyword.

Required flags:
  --url-id        ID of the tracked URL
  --keyword-id    ID of the keyword to update
  --tags          Comma-separated list of tags

Example:
  nightwatch keywords update --url-id 123 --keyword-id 456 --tags "priority,brand"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlID, _ := cmd.Flags().GetString("url-id")
		if urlID == "" {
			output.PrintErrorTyped(os.Stderr, "--url-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		keywordID, _ := cmd.Flags().GetString("keyword-id")
		if keywordID == "" {
			output.PrintErrorTyped(os.Stderr, "--keyword-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		tagsStr, _ := cmd.Flags().GetString("tags")
		if tagsStr == "" {
			output.PrintErrorTyped(os.Stderr, "--tags is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		tags := strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}

		endpoint := fmt.Sprintf("/urls/%s/keywords/%s", urlID, keywordID)
		body := map[string]any{
			"keyword": map[string]any{
				"tags": tags,
			},
		}

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

func init() {
	// list flags
	keywordsListCmd.Flags().String("url-id", "", "ID of the tracked URL (required)")
	keywordsListCmd.Flags().String("view-id", "", "Filter by dynamic view ID")
	keywordsListCmd.Flags().String("page", "", "Page number")
	keywordsListCmd.Flags().String("limit", "", "Results per page")
	keywordsListCmd.Flags().String("search", "", "Search keywords by query")
	keywordsListCmd.Flags().String("sort", "", "Sort field (e.g. position, query)")
	keywordsListCmd.Flags().String("direction", "", "Sort direction: asc or desc")

	// add flags
	keywordsAddCmd.Flags().String("url-id", "", "ID of the tracked URL (required)")
	keywordsAddCmd.Flags().String("keywords", "", "Comma-separated keywords to add (required)")
	keywordsAddCmd.Flags().String("google-gl", "", "Google geolocation code, e.g. us, gb (required)")
	keywordsAddCmd.Flags().String("google-hl", "", "Google language code, e.g. en, de (required)")
	keywordsAddCmd.Flags().String("search-engine", "", "Search engine: google, bing, youtube (required)")
	keywordsAddCmd.Flags().Bool("mobile", false, "Track mobile rankings")
	keywordsAddCmd.Flags().Bool("desktop", false, "Track desktop rankings")
	keywordsAddCmd.Flags().String("tags", "", "Comma-separated tags to assign")
	keywordsAddCmd.Flags().String("adwords-location-id", "", "Google Ads location ID for local results")

	// remove flags
	keywordsRemoveCmd.Flags().String("url-id", "", "ID of the tracked URL (required)")
	keywordsRemoveCmd.Flags().String("keyword-ids", "", "Comma-separated keyword IDs to remove (required)")

	// update flags
	keywordsUpdateCmd.Flags().String("url-id", "", "ID of the tracked URL (required)")
	keywordsUpdateCmd.Flags().String("keyword-id", "", "ID of the keyword to update (required)")
	keywordsUpdateCmd.Flags().String("tags", "", "Comma-separated tags (required)")

	keywordsCmd.AddCommand(keywordsListCmd)
	keywordsCmd.AddCommand(keywordsAddCmd)
	keywordsCmd.AddCommand(keywordsRemoveCmd)
	keywordsCmd.AddCommand(keywordsUpdateCmd)
	rootCmd.AddCommand(keywordsCmd)
}
