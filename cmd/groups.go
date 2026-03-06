package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nightwatch-io/nightwatch-cli/internal/client"
	"github.com/nightwatch-io/nightwatch-cli/internal/output"
	"github.com/spf13/cobra"
)

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage URL groups for organizing tracked URLs",
	Long: `Manage URL groups in your Nightwatch account. Groups let you organize
tracked URLs into logical categories.

Available subcommands:
  list     List all URL groups
  get      Get details for a specific group by ID
  create   Create a new URL group
  update   Update an existing group's name
  delete   Delete a URL group

Every response is a JSON envelope: {"data": ..., "meta": {...}}`,
	Annotations: map[string]string{AnnotationRequiresAuth: "true"},
}

var groupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all URL groups",
	Long: `Retrieve all URL groups in your Nightwatch account.

Example:
  nightwatch groups list
  nightwatch groups list --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun(cmd) {
			return printDryRunGet(cmd, "/url_groups", nil)
		}

		cl, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := cl.Get(cmd.Context(), "/url_groups", nil)
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

var groupsGetCmd = &cobra.Command{
	Use:   "get ID",
	Short: "Get details for a specific URL group",
	Long: `Retrieve details for a single URL group by its numeric ID.

Example:
  nightwatch groups get 123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		endpoint := fmt.Sprintf("/url_groups/%s", id)

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

var groupsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new URL group",
	Long: `Create a new URL group with the given name.

Example:
  nightwatch groups create --name "My SEO Projects"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			output.PrintErrorTyped(os.Stderr, "--name is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		body := map[string]any{
			"url_group": map[string]any{
				"name": name,
			},
		}

		if isDryRun(cmd) {
			return printDryRunPost(cmd, "/url_groups", body)
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

		resp, err := cl.Post(cmd.Context(), "/url_groups", jsonBody)
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

var groupsUpdateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update an existing URL group",
	Long: `Update the name of an existing URL group by its numeric ID.

Example:
  nightwatch groups update 123 --name "Renamed Group"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			output.PrintErrorTyped(os.Stderr, "--name is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		endpoint := fmt.Sprintf("/url_groups/%s", id)
		body := map[string]any{
			"url_group": map[string]any{
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

		data, err := parseResponseData(resp.Body)
		if err != nil {
			output.PrintErrorTyped(os.Stderr, err.Error(), 1, client.ErrorTypeClient)
			return &exitError{code: 1}
		}

		output.PrintSingle(cmd.OutOrStdout(), data)
		return nil
	},
}

var groupsDeleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "Delete a URL group",
	Long: `Delete a URL group by its numeric ID.

Example:
  nightwatch groups delete 123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		endpoint := fmt.Sprintf("/url_groups/%s", id)

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
	groupsCreateCmd.Flags().String("name", "", "Name for the new URL group (required)")
	groupsUpdateCmd.Flags().String("name", "", "New name for the URL group (required)")

	groupsCmd.AddCommand(groupsListCmd)
	groupsCmd.AddCommand(groupsGetCmd)
	groupsCmd.AddCommand(groupsCreateCmd)
	groupsCmd.AddCommand(groupsUpdateCmd)
	groupsCmd.AddCommand(groupsDeleteCmd)
	rootCmd.AddCommand(groupsCmd)
}
