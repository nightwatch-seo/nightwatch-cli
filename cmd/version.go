package cmd

import (
	"github.com/nightwatch-io/nightwatch-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

// SetVersionInfo stores build-time version metadata for the version command.
func SetVersionInfo(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
	rootCmd.Version = version
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print CLI version, git commit SHA, and build date as JSON",
	Long: `Print the CLI build version, git commit SHA, and build date as a JSON object.

Example:
  nightwatch version

Output:
  {"data":{"version":"0.1.0","commit":"abc123","date":"2026-03-01"},"meta":{}}`,
	Run: func(cmd *cobra.Command, args []string) {
		data := map[string]any{
			"version": buildVersion,
			"commit":  buildCommit,
			"date":    buildDate,
		}
		output.PrintData(cmd.OutOrStdout(), data)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
