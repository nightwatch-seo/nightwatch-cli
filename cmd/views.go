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

var viewsCmd = &cobra.Command{
	Use:   "views",
	Short: "Manage keyword views (dynamic filtered views)",
	Long: `Manage keyword views in your Nightwatch account. Views are dynamic filters
that segment keywords based on rules (e.g. position ranges, tags, query matches).

Available subcommands:
  list         List keyword views
  get          Get details for a specific view
  create       Create a new keyword view
  update       Update a view's name
  delete       Delete a keyword view
  add-filter   Add a filter group to a view

Every response is a JSON envelope: {"data": ..., "meta": {...}}`,
	Annotations: map[string]string{AnnotationRequiresAuth: "true"},
}

var viewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List keyword views",
	Long: `List keyword views, optionally filtered by URL or group.

Optional flags:
  --url-id           Filter views by URL ID
  --group-id         Filter views by group ID
  --without-counts   Exclude keyword counts from response

Examples:
  nightwatch views list
  nightwatch views list --url-id 123
  nightwatch views list --group-id 42 --without-counts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("url-id"); v != "" {
			q.Set("url_id", v)
		}
		if v, _ := cmd.Flags().GetString("group-id"); v != "" {
			q.Set("group_id", v)
		}
		if cmd.Flags().Changed("without-counts") {
			v, _ := cmd.Flags().GetBool("without-counts")
			if v {
				q.Set("without_counts", "true")
			}
		}

		if isDryRun(cmd) {
			return printDryRunGet(cmd, "/dynamic_views", q)
		}

		cl, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := cl.Get(cmd.Context(), "/dynamic_views", q)
		if err != nil {
			return handleAPIError(err)
		}

		return printResponse(cmd, resp.Body, printAsList)
	},
}

var viewsGetCmd = &cobra.Command{
	Use:   "get ID",
	Short: "Get details for a specific keyword view",
	Long: `Retrieve details for a single keyword view by its numeric ID.

Example:
  nightwatch views get 789`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		endpoint := fmt.Sprintf("/dynamic_views/%s", id)

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

		return printResponse(cmd, resp.Body, printAsSingle)
	},
}

var viewsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new keyword view",
	Long: `Create a new keyword view. Must be scoped to either a URL or a group.

Required flags:
  --name       Name for the view

Scope (at least one required):
  --url-id     Scope the view to a specific URL
  --group-id   Scope the view to a URL group

Examples:
  nightwatch views create --name "Top 10 Keywords" --url-id 123
  nightwatch views create --name "Brand Keywords" --group-id 42`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			output.PrintErrorTyped(os.Stderr, "--name is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		urlID, _ := cmd.Flags().GetString("url-id")
		groupID, _ := cmd.Flags().GetString("group-id")
		if urlID == "" && groupID == "" {
			output.PrintErrorTyped(os.Stderr, "one of --url-id or --group-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		viewData := map[string]any{
			"name": name,
		}
		if urlID != "" {
			viewData["search_keyword_url_id"] = urlID
		}
		if groupID != "" {
			viewData["url_group_id"] = groupID
		}

		body := map[string]any{"dynamic_view": viewData}

		if isDryRun(cmd) {
			return printDryRunPost(cmd, "/dynamic_views", body)
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

		resp, err := cl.Post(cmd.Context(), "/dynamic_views", jsonBody)
		if err != nil {
			return handleAPIError(err)
		}

		return printResponse(cmd, resp.Body, printAsSingle)
	},
}

var viewsUpdateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update a keyword view's name",
	Long: `Update the name of an existing keyword view by its numeric ID.

Example:
  nightwatch views update 789 --name "Renamed View"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			output.PrintErrorTyped(os.Stderr, "--name is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		endpoint := fmt.Sprintf("/dynamic_views/%s", id)
		body := map[string]any{
			"dynamic_view": map[string]any{
				"name": name,
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

		return printResponse(cmd, resp.Body, printAsSingle)
	},
}

var viewsDeleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "Delete a keyword view",
	Long: `Delete a keyword view by its numeric ID.

Example:
  nightwatch views delete 789`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		endpoint := fmt.Sprintf("/dynamic_views/%s", id)

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

var viewsAddFilterCmd = &cobra.Command{
	Use:   "add-filter ID",
	Short: "Add a filter group to a keyword view",
	Long: `Add a filter group to an existing keyword view. Filters define rules
for which keywords appear in the view.

The --filters flag accepts a JSON array of filter objects. Each filter has:
  - field:      The keyword field to filter on (e.g. "query", "position", "tags")
  - condition:  The comparison operator (e.g. "contains", "equals", "greater_than")
  - value:      The value to compare against (optional for some conditions)

Examples:
  nightwatch views add-filter 789 --filters '[{"field":"query","condition":"contains","value":"seo"}]'
  nightwatch views add-filter 789 --filters '[{"field":"position","condition":"less_than","value":"10"}]'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		filtersStr, _ := cmd.Flags().GetString("filters")
		if filtersStr == "" {
			output.PrintErrorTyped(os.Stderr, "--filters is required (JSON array)", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		var filters []any
		if err := json.Unmarshal([]byte(filtersStr), &filters); err != nil {
			output.PrintErrorTyped(os.Stderr, "invalid --filters JSON: "+err.Error(), 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		endpoint := fmt.Sprintf("/dynamic_views/%s/filter_groups", id)
		body := map[string]any{
			"filter_group": map[string]any{
				"filters": filters,
			},
		}

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

		return printResponse(cmd, resp.Body, printAsSingle)
	},
}

func init() {
	// list flags
	viewsListCmd.Flags().String("url-id", "", "Filter views by URL ID")
	viewsListCmd.Flags().String("group-id", "", "Filter views by group ID")
	viewsListCmd.Flags().Bool("without-counts", false, "Exclude keyword counts from response")

	// create flags
	viewsCreateCmd.Flags().String("name", "", "Name for the view (required)")
	viewsCreateCmd.Flags().String("url-id", "", "Scope the view to a specific URL")
	viewsCreateCmd.Flags().String("group-id", "", "Scope the view to a URL group")

	// update flags
	viewsUpdateCmd.Flags().String("name", "", "New name for the view (required)")

	// add-filter flags
	viewsAddFilterCmd.Flags().String("filters", "", "JSON array of filter objects (required)")

	viewsCmd.AddCommand(viewsListCmd)
	viewsCmd.AddCommand(viewsGetCmd)
	viewsCmd.AddCommand(viewsCreateCmd)
	viewsCmd.AddCommand(viewsUpdateCmd)
	viewsCmd.AddCommand(viewsDeleteCmd)
	viewsCmd.AddCommand(viewsAddFilterCmd)
	rootCmd.AddCommand(viewsCmd)
}
