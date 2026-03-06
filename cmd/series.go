package cmd

import (
	"net/url"
	"os"
	"strings"

	"github.com/nightwatch-io/nightwatch-cli/internal/client"
	"github.com/nightwatch-io/nightwatch-cli/internal/output"
	"github.com/spf13/cobra"
)

var seriesCmd = &cobra.Command{
	Use:   "series",
	Short: "Retrieve historical ranking data series",
	Long: `Retrieve historical keyword ranking data series from Nightwatch.

Available subcommands:
  get    Get historical data for keywords, URLs, views, or groups

Every response is a JSON envelope: {"data": ..., "meta": {...}}`,
	Annotations: map[string]string{AnnotationRequiresAuth: "true"},
}

var seriesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get historical ranking data series",
	Long: `Retrieve historical ranking data for a date range. At least one of
--url-ids, --keyword-ids, --view-ids, or --group-ids must be provided.

Required flags:
  --date-from    Start date in YYYY-MM-DD format
  --date-to      End date in YYYY-MM-DD format

Required (at least one):
  --url-ids        Comma-separated URL IDs
  --keyword-ids    Comma-separated keyword IDs
  --view-ids       Comma-separated view IDs
  --group-ids      Comma-separated group IDs

Optional flags:
  --with-competitors    Include competitor data in the series

Examples:
  nightwatch series get --date-from 2025-01-01 --date-to 2025-03-01 --url-ids 123
  nightwatch series get --date-from 2025-01-01 --date-to 2025-03-01 --keyword-ids 456,789 --with-competitors`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dateFrom, _ := cmd.Flags().GetString("date-from")
		dateTo, _ := cmd.Flags().GetString("date-to")

		if dateFrom == "" {
			output.PrintErrorTyped(os.Stderr, "--date-from is required (YYYY-MM-DD)", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}
		if dateTo == "" {
			output.PrintErrorTyped(os.Stderr, "--date-to is required (YYYY-MM-DD)", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		q := url.Values{}
		q.Set("date_from", dateFrom)
		q.Set("date_to", dateTo)

		hasIDs := false
		if v, _ := cmd.Flags().GetString("url-ids"); v != "" {
			for _, id := range strings.Split(v, ",") {
				q.Add("url_ids[]", strings.TrimSpace(id))
			}
			hasIDs = true
		}
		if v, _ := cmd.Flags().GetString("keyword-ids"); v != "" {
			for _, id := range strings.Split(v, ",") {
				q.Add("keyword_ids[]", strings.TrimSpace(id))
			}
			hasIDs = true
		}
		if v, _ := cmd.Flags().GetString("view-ids"); v != "" {
			for _, id := range strings.Split(v, ",") {
				q.Add("dynamic_view_ids[]", strings.TrimSpace(id))
			}
			hasIDs = true
		}
		if v, _ := cmd.Flags().GetString("group-ids"); v != "" {
			for _, id := range strings.Split(v, ",") {
				q.Add("url_group_ids[]", strings.TrimSpace(id))
			}
			hasIDs = true
		}

		if !hasIDs {
			output.PrintErrorTyped(os.Stderr, "at least one of --url-ids, --keyword-ids, --view-ids, or --group-ids is required", 1, client.ErrorTypeValidation)
			return &exitError{code: 1}
		}

		if cmd.Flags().Changed("with-competitors") {
			v, _ := cmd.Flags().GetBool("with-competitors")
			if v {
				q.Set("with_competitors", "true")
			}
		}

		if isDryRun(cmd) {
			return printDryRunGet(cmd, "/series", q)
		}

		cl, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}

		resp, err := cl.Get(cmd.Context(), "/series", q)
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
	seriesGetCmd.Flags().String("date-from", "", "Start date in YYYY-MM-DD format (required)")
	seriesGetCmd.Flags().String("date-to", "", "End date in YYYY-MM-DD format (required)")
	seriesGetCmd.Flags().String("url-ids", "", "Comma-separated URL IDs")
	seriesGetCmd.Flags().String("keyword-ids", "", "Comma-separated keyword IDs")
	seriesGetCmd.Flags().String("view-ids", "", "Comma-separated view IDs")
	seriesGetCmd.Flags().String("group-ids", "", "Comma-separated group IDs")
	seriesGetCmd.Flags().Bool("with-competitors", false, "Include competitor data in the series")

	seriesCmd.AddCommand(seriesGetCmd)
	rootCmd.AddCommand(seriesCmd)
}
