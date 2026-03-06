package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serpsCmd = &cobra.Command{
	Use:   "serps",
	Short: "Retrieve SERP (search engine results page) data",
	Long: `Retrieve SERP data from Nightwatch for specific keyword results.

Available subcommands:
  get    Get SERP data for a specific keyword result ID

Every response is a JSON envelope: {"data": ..., "meta": {...}}`,
	Annotations: map[string]string{AnnotationRequiresAuth: "true"},
}

var serpsGetCmd = &cobra.Command{
	Use:   "get ID",
	Short: "Get SERP data for a keyword result",
	Long: `Retrieve the SERP (search engine results page) data for a specific
keyword result ID. This shows the full search results page that was
captured for the keyword at a specific point in time.

Example:
  nightwatch serps get 12345`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		endpoint := fmt.Sprintf("/serp_data/%s", id)

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

func init() {
	serpsCmd.AddCommand(serpsGetCmd)
	rootCmd.AddCommand(serpsCmd)
}
