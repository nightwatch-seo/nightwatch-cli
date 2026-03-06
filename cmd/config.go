package cmd

import (
	"os"

	"github.com/nightwatch-io/nightwatch-cli/internal/client"
	"github.com/nightwatch-io/nightwatch-cli/internal/config"
	"github.com/nightwatch-io/nightwatch-cli/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Read and write persistent CLI settings (API key, base URL)",
	Long: `Manage persistent CLI configuration stored in ~/.config/nightwatch/config.yaml.

Available subcommands:
  set    Save a configuration key (api-key, base-url) to the config file
  show   Display the resolved configuration as JSON (API key is masked)`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display the resolved configuration as JSON (API key masked)",
	Long: `Print the resolved configuration as a JSON object to stdout. The API key
is masked for security (first 4 and last 4 characters shown). Values
reflect the full precedence chain: NIGHTWATCH_API_KEY env > config file > default.

Example:
  nightwatch config show`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := ResolvedConfig()
		data := map[string]any{
			"api_key":  config.MaskAPIKey(cfg.APIKey),
			"base_url": cfg.BaseURL,
		}
		output.PrintData(cmd.OutOrStdout(), data)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set KEY VALUE",
	Short: "Save a configuration key to ~/.config/nightwatch/config.yaml",
	Long: `Write a configuration key-value pair to ~/.config/nightwatch/config.yaml.

Supported keys:
  api-key     Your Nightwatch API token
  base-url    API base URL (default: https://api.nightwatch.io/api/v1)

Examples:
  Save an API key:
    nightwatch config set api-key your-api-token-here

  Override the base URL:
    nightwatch config set base-url https://api.example.com/v1`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		yamlKey, ok := keyMapping[key]
		if !ok {
			msg := "unknown config key: " + key + ". Supported keys: api-key, base-url"
			output.PrintErrorTyped(os.Stderr, msg, 1, client.ErrorTypeClient)
			return &exitError{code: 1}
		}

		if err := config.SaveToFile(yamlKey, value); err != nil {
			output.PrintErrorTyped(os.Stderr, err.Error(), 1, client.ErrorTypeClient)
			return &exitError{code: 1}
		}

		output.PrintData(cmd.OutOrStdout(), map[string]string{
			"status": "ok",
			"key":    key,
		})
		return nil
	},
}

// keyMapping translates CLI-friendly key names to YAML config field names.
var keyMapping = map[string]string{
	"api-key":  "api_key",
	"base-url": "base_url",
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
