package cmd

import (
	"net/url"
	"os"

	"github.com/nightwatch-io/nightwatch-cli/internal/client"
	"github.com/nightwatch-io/nightwatch-cli/internal/output"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get keyword statistics for date ranges",
	Long: `Retrieve keyword statistics calculated over a date range.

Available subcommands:
  get    Get keyword statistics

Every response is a JSON envelope: {"data": ..., "meta": {...}}`,
	Annotations: map[string]string{AnnotationRequiresAuth: "true"},
}

var statsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get keyword statistics for a date range",
	Long: `Calculate keyword statistics over a specified date range. Must be scoped
to either a URL or a dynamic view.

Required flags:
  --start-date    Start date in YYYY-MM-DD format
  --end-date      End date in YYYY-MM-DD format

Required (one of):
  --url-id        Scope stats to a specific URL
  --view-id       Scope stats to a dynamic view

Optional flags:
  --include-organic    Include organic search metrics

Examples:
  nightwatch stats get --start-date 2025-01-01 --end-date 2025-03-01 --url-id 123
  nightwatch stats get --start-date 2025-01-01 --end-date 2025-03-01 --view-id 789 --include-organic`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")

		if startDate == "" {
			output.PrintErrorTyped(os.Stderr, "--start-date is required (YYYY-MM-DD)", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}
		if err := validateDate("--start-date", startDate); err != nil {
			return err
		}
		if endDate == "" {
			output.PrintErrorTyped(os.Stderr, "--end-date is required (YYYY-MM-DD)", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}
		if err := validateDate("--end-date", endDate); err != nil {
			return err
		}

		urlID, _ := cmd.Flags().GetString("url-id")
		viewID, _ := cmd.Flags().GetString("view-id")
		if urlID == "" && viewID == "" {
			output.PrintErrorTyped(os.Stderr, "one of --url-id or --view-id is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		q := url.Values{}
		q.Set("start_date", startDate)
		q.Set("end_date", endDate)

		if urlID != "" {
			q.Set("url_id", urlID)
		}
		if viewID != "" {
			q.Set("dynamic_view_id", viewID)
		}
		if cmd.Flags().Changed("include-organic") {
			v, _ := cmd.Flags().GetBool("include-organic")
			if v {
				q.Set("include_organic", "true")
			}
		}

		if isDryRun(cmd) {
			return printDryRunGet(cmd, "/keyword_stats", q)
		}

		cl, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := cl.Get(cmd.Context(), "/keyword_stats", q)
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
	statsGetCmd.Flags().String("start-date", "", "Start date in YYYY-MM-DD format (required)")
	statsGetCmd.Flags().String("end-date", "", "End date in YYYY-MM-DD format (required)")
	statsGetCmd.Flags().String("url-id", "", "Scope stats to a specific URL")
	statsGetCmd.Flags().String("view-id", "", "Scope stats to a dynamic view")
	statsGetCmd.Flags().Bool("include-organic", false, "Include organic search metrics")

	statsCmd.AddCommand(statsGetCmd)
	rootCmd.AddCommand(statsCmd)
}
