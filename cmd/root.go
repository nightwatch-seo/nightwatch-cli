package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/nightwatch-io/nightwatch-cli/internal/client"
	"github.com/nightwatch-io/nightwatch-cli/internal/config"
	"github.com/nightwatch-io/nightwatch-cli/internal/output"
	"github.com/spf13/cobra"
)

// AnnotationRequiresAuth is the cobra annotation key used to mark commands
// that need a valid API key. Commands without this annotation skip auth checks.
const AnnotationRequiresAuth = "requires_auth"

// resolvedConfig holds the config resolved during PersistentPreRun, available
// to all subcommands within the same execution.
var resolvedConfig config.Config

// ResolvedConfig returns the config resolved during PersistentPreRun.
func ResolvedConfig() config.Config {
	return resolvedConfig
}

// flagErrPrinted tracks whether SetFlagErrorFunc already emitted JSON to
// stderr, preventing Execute from printing a duplicate error.
var flagErrPrinted bool

// exitError carries a specific exit code through cobra's error chain.
type exitError struct {
	code int
}

func (e *exitError) Error() string { return "" }

var rootCmd = &cobra.Command{
	Use:   "nightwatch",
	Short: "Query the Nightwatch.io SEO API from the command line",
	Long: `nightwatch is a CLI for the Nightwatch.io REST API. It manages SEO keyword
tracking, URL monitoring, competitor analysis, and SERP data retrieval.

Every command outputs valid JSON to stdout. Errors go to stderr as JSON.
Pipe output to jq, parse it programmatically, or feed it to another AI agent.

Available resources:
  groups        Manage URL groups for organizing tracked URLs
  urls          Manage tracked URLs (projects) and their settings
  keywords      Add, remove, update, and list tracked keywords for a URL
  views         Manage keyword views (dynamic filtered views)
  competitors   Track and manage competitor URLs
  series        Retrieve historical ranking data series
  stats         Get keyword statistics for date ranges
  serps         Retrieve SERP (search engine results page) data
  config        Read and write persistent settings (API key, base URL)
  version       Print CLI version, git commit, and build date

Authentication (checked in this order):
  1. NIGHTWATCH_API_KEY environment variable
  2. ~/.config/nightwatch/config.yaml

Quick start:
  export NIGHTWATCH_API_KEY=your-token-here
  # or: nightwatch config set api-key YOUR_TOKEN
  nightwatch groups list
  nightwatch urls list
  nightwatch keywords list --url-id 123`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		flagBaseURL, _ := cmd.Flags().GetString("base-url")

		o := config.Overrides{
			BaseURL:    flagBaseURL,
			BaseURLSet: cmd.Flags().Changed("base-url"),
		}

		cfg, err := config.Resolve(o)
		if err != nil {
			if cmd.Name() == "version" {
				resolvedConfig = config.FallbackConfig(o)
				return nil
			}
			output.PrintErrorTyped(os.Stderr, err.Error(), 1, client.ErrorTypeClient)
			return &exitError{code: 1}
		}
		resolvedConfig = cfg

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if requiresAuth(cmd) && cfg.APIKey == "" && !dryRun {
			msg := "API key not configured. Set NIGHTWATCH_API_KEY or run: nightwatch config set api-key <token>"
			output.PrintErrorTyped(os.Stderr, msg, 2, client.ErrorTypeAuth)
			return &exitError{code: 2}
		}

		return nil
	},
}

// requiresAuth checks whether the command (or any of its parents) is
// annotated as requiring authentication.
func requiresAuth(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Annotations != nil {
			if _, ok := c.Annotations[AnnotationRequiresAuth]; ok {
				return true
			}
		}
	}
	return false
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.String("base-url", config.DefaultBaseURL, "API base URL")
	flags.Bool("no-retry", false, "Disable automatic retry on HTTP 429 rate-limit responses")
	flags.String("timeout", "30s", "Per-request timeout as a Go duration, e.g. 10s, 1m, 2m30s")
	flags.Bool("dry-run", false, "Print the resolved HTTP request as JSON without calling the API")

	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		flagErrPrinted = true
		output.PrintErrorTyped(os.Stderr, err.Error(), 1, client.ErrorTypeClient)
		return err
	})
}

// Execute runs the root command and returns the exit code.
func Execute() int {
	flagErrPrinted = false
	if err := rootCmd.Execute(); err != nil {
		if e, ok := err.(*exitError); ok {
			return e.code
		}
		if !flagErrPrinted {
			output.PrintErrorTyped(os.Stderr, err.Error(), 1, client.ErrorTypeClient)
		}
		return 1
	}
	return 0
}

// newClientFromFlags creates a client.Client from the resolved config and
// the --timeout / --no-retry flags.
func newClientFromFlags(cmd *cobra.Command) (*client.Client, error) {
	cfg := ResolvedConfig()

	timeout, err := validateTimeout(cmd)
	if err != nil {
		return nil, err
	}

	noRetry, _ := cmd.Flags().GetBool("no-retry")

	cl := client.New(client.Options{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Timeout: timeout,
		NoRetry: noRetry,
		Version: buildVersion,
	})
	return cl, nil
}

// validateTimeout parses the --timeout flag and returns the duration.
func validateTimeout(cmd *cobra.Command) (time.Duration, error) {
	raw, _ := cmd.Flags().GetString("timeout")
	d, err := time.ParseDuration(raw)
	if err != nil {
		msg := fmt.Sprintf("invalid --timeout value %q: %v", raw, err)
		output.PrintErrorTyped(os.Stderr, msg, 1, client.ErrorTypeClient)
		return 0, &exitError{code: 1}
	}
	return d, nil
}

// isDryRun returns true when the --dry-run flag is set on the command.
func isDryRun(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("dry-run")
	return v
}

// dryRunRequest is the JSON structure printed when --dry-run is active.
type dryRunRequest struct {
	Method string         `json:"method"`
	URL    string         `json:"url"`
	Params map[string]any `json:"params,omitempty"`
	Body   any            `json:"body,omitempty"`
}

// printDryRunGet writes the resolved GET request to stdout.
func printDryRunGet(cmd *cobra.Command, endpoint string, query url.Values) error {
	cfg := ResolvedConfig()
	fullURL := cfg.BaseURL + endpoint

	out := dryRunRequest{
		Method: "GET",
		URL:    fullURL,
		Params: valuesToMap(query),
	}

	return encodeDryRun(cmd, out)
}

// printDryRunPost writes the resolved POST request to stdout.
func printDryRunPost(cmd *cobra.Command, endpoint string, body any) error {
	cfg := ResolvedConfig()
	fullURL := cfg.BaseURL + endpoint

	out := dryRunRequest{
		Method: "POST",
		URL:    fullURL,
		Body:   body,
	}

	return encodeDryRun(cmd, out)
}

// printDryRunPut writes the resolved PUT request to stdout.
func printDryRunPut(cmd *cobra.Command, endpoint string, body any) error {
	cfg := ResolvedConfig()
	fullURL := cfg.BaseURL + endpoint

	out := dryRunRequest{
		Method: "PUT",
		URL:    fullURL,
		Body:   body,
	}

	return encodeDryRun(cmd, out)
}

// printDryRunDelete writes the resolved DELETE request to stdout.
func printDryRunDelete(cmd *cobra.Command, endpoint string) error {
	cfg := ResolvedConfig()
	fullURL := cfg.BaseURL + endpoint

	out := dryRunRequest{
		Method: "DELETE",
		URL:    fullURL,
	}

	return encodeDryRun(cmd, out)
}

func encodeDryRun(cmd *cobra.Command, out dryRunRequest) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("failed to encode dry-run output: %w", err)
	}
	return nil
}

// valuesToMap converts url.Values into a map suitable for JSON output.
func valuesToMap(query url.Values) map[string]any {
	if len(query) == 0 {
		return nil
	}
	m := make(map[string]any, len(query))
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := query[k]
		if len(v) == 1 {
			m[k] = v[0]
		} else {
			m[k] = v
		}
	}
	return m
}

// handleAPIError translates an error from the client into a structured
// JSON error on stderr and returns an exitError.
func handleAPIError(err error) error {
	apiErr, ok := err.(*client.APIError)
	if ok {
		output.PrintErrorTyped(os.Stderr, apiErr.Message, apiErr.ExitCode, apiErr.ErrorType)
		return &exitError{code: apiErr.ExitCode}
	}
	output.PrintErrorTyped(os.Stderr, err.Error(), 1, client.ErrorTypeClient)
	return &exitError{code: 1}
}

// parseResponseData unmarshals the API response body into an interface.
func parseResponseData(body []byte) (any, error) {
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}
	return data, nil
}

// validateDate checks that a date string is in YYYY-MM-DD format and
// represents a valid calendar date. Returns a validation exitError on failure.
func validateDate(flagName, value string) error {
	if _, err := time.Parse("2006-01-02", value); err != nil {
		msg := fmt.Sprintf("invalid %s %q: expected YYYY-MM-DD format", flagName, value)
		output.PrintErrorTyped(os.Stderr, msg, 1, client.ErrorTypeValidation)
		return &exitError{code: 1}
	}
	return nil
}
