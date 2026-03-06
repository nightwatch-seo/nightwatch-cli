# Nightwatch CLI

Agent-first CLI for the [Nightwatch.io](https://nightwatch.io/) SEO keyword tracking API. A single binary that any AI agent (or human) can shell out to. JSON only. Zero interactivity.

## What is this

[Nightwatch](https://nightwatch.io/) tracks keyword rankings, SERP data, and SEO performance across search engines. This CLI wraps that API so you can query and manage it from the command line.

**Designed for AI agents.** Every command prints valid JSON to stdout and structured JSON errors to stderr. No prompts, spinners, colors, or interactive elements. Help text is written for LLMs. Exit codes are meaningful and documented.

**Designed for scripts.** Pipe output to `jq`, feed it to another process, or parse it in any language.

**What you can do:**
- Manage URL groups, tracked URLs, and keyword lists
- Track keyword rankings across Google, Bing, YouTube
- Create dynamic keyword views with filters
- Monitor competitors and compare rankings
- Retrieve historical ranking series and statistics
- Access SERP data for keyword results

Get an API token at [nightwatch.io](https://nightwatch.io/) to get started.

## Install

```bash
go build -o nightwatch .
```

## Quick start

```bash
# 1. Set your API token
export NIGHTWATCH_API_KEY=your-token

# 2. List your URL groups
nightwatch groups list

# 3. List tracked URLs
nightwatch urls list

# 4. List keywords for a URL
nightwatch keywords list --url-id 123

# 5. Get ranking history
nightwatch series get --date-from 2025-01-01 --date-to 2025-03-01 --url-ids 123
```

Or save the API token to the config file:

```bash
nightwatch config set api-key your-token
```

## Authentication

| Priority | Source | Example |
|----------|--------|---------|
| 1 | `NIGHTWATCH_API_KEY` env var | `export NIGHTWATCH_API_KEY=abc123` |
| 2 | Config file | `~/.config/nightwatch/config.yaml` |

## Commands

```
nightwatch
├── groups
│   ├── list        List all URL groups
│   ├── get         Get a group by ID
│   ├── create      Create a new group
│   ├── update      Update a group's name
│   └── delete      Delete a group
├── urls
│   ├── list        List tracked URLs (optionally by group)
│   ├── get         Get a URL by ID
│   ├── create      Add a new URL to track
│   ├── update      Update URL settings
│   └── delete      Remove a tracked URL
├── keywords
│   ├── list        List keywords for a URL (with pagination/filtering)
│   ├── add         Batch-add keywords to a URL
│   ├── remove      Batch-remove keywords by ID
│   └── update      Update keyword tags
├── views
│   ├── list        List keyword views
│   ├── get         Get a view by ID
│   ├── create      Create a new view
│   ├── update      Update a view's name
│   ├── delete      Delete a view
│   └── add-filter  Add a filter group to a view
├── competitors
│   ├── list        List competitors for a URL
│   ├── add         Add a competitor
│   └── remove      Remove a competitor
├── series
│   └── get         Get historical ranking data
├── stats
│   └── get         Get keyword statistics for a date range
├── serps
│   └── get         Get SERP data for a keyword result
├── config
│   ├── set         Save a configuration value
│   └── show        Display resolved configuration (API key masked)
├── version         Print CLI version, git commit, and build date
├── completion      Generate shell autocompletion scripts
└── help            Help about any command
```

### groups

Manage URL groups for organizing tracked URLs.

```bash
nightwatch groups list
nightwatch groups get 42
nightwatch groups create --name "My Projects"
nightwatch groups update 42 --name "Renamed"
nightwatch groups delete 42
```

### urls

Manage tracked URLs (projects).

```bash
nightwatch urls list
nightwatch urls list --group-id 42
nightwatch urls create --url "https://example.com" --group-id 42 --country-code us --language-code en
nightwatch urls update 123 --custom-name "Main Site"
nightwatch urls delete 123
```

### keywords

Manage keywords tracked for a URL.

```bash
# List with pagination and filtering
nightwatch keywords list --url-id 123 --page 1 --limit 50 --search "seo"

# Batch-add keywords
nightwatch keywords add --url-id 123 --keywords "seo tools,keyword tracking" \
  --google-gl us --google-hl en --search-engine google

# Batch-remove by ID
nightwatch keywords remove --url-id 123 --keyword-ids 456,789

# Update tags
nightwatch keywords update --url-id 123 --keyword-id 456 --tags "brand,priority"
```

### views

Manage keyword views (dynamic filtered views).

```bash
nightwatch views list --url-id 123
nightwatch views create --name "Top 10" --url-id 123
nightwatch views add-filter 789 --filters '[{"field":"query","condition":"contains","value":"seo"}]'
nightwatch views delete 789
```

### competitors

Track competitor URLs.

```bash
nightwatch competitors list --url-id 123
nightwatch competitors add --url-id 123 --url "https://competitor.com" --custom-name "Main Rival"
nightwatch competitors remove --url-id 123 --competitor-id 456
```

### series

Get historical ranking data.

```bash
nightwatch series get --date-from 2025-01-01 --date-to 2025-03-01 --url-ids 123
nightwatch series get --date-from 2025-01-01 --date-to 2025-03-01 --keyword-ids 456,789 --with-competitors
```

### stats

Get keyword statistics for a date range.

```bash
nightwatch stats get --start-date 2025-01-01 --end-date 2025-03-01 --url-id 123
nightwatch stats get --start-date 2025-01-01 --end-date 2025-03-01 --view-id 789 --include-organic
```

### serps

Get SERP data for a keyword result.

```bash
nightwatch serps get 12345
```

### config

```bash
nightwatch config set api-key your-token
nightwatch config show
```

Settings are stored in `~/.config/nightwatch/config.yaml`.

### version

```bash
nightwatch version
```

## Output format

Every command writes valid JSON to **stdout**. Errors go to **stderr** as JSON. stdout is always parseable.

**Success:**
```json
{
  "data": {...},
  "meta": {}
}
```

**Error (stderr):**
```json
{
  "error": "API key not configured. Set NIGHTWATCH_API_KEY or run: nightwatch config set api-key <token>",
  "code": 2,
  "error_type": "auth_error"
}
```

Possible `error_type` values: `client_error`, `validation_error`, `auth_error`, `rate_limited`, `server_error`, `network_error`.

## Dry run

Preview any request without calling the API:

```bash
nightwatch groups list --dry-run
nightwatch urls create --url "https://example.com" --group-id 42 --country-code us --language-code en --dry-run
```

## Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--base-url` | `https://api.nightwatch.io/api/v1` | API base URL |
| `--no-retry` | `false` | Disable retry on HTTP 429 |
| `--timeout` | `30s` | Per-request timeout |
| `--dry-run` | `false` | Print request without calling API |

## Exit codes

| Code | Meaning | `error_type` |
|------|---------|--------------|
| 0 | Success | — |
| 1 | Client error | `client_error`, `validation_error` |
| 2 | Auth error | `auth_error` |
| 3 | Rate limited | `rate_limited` |
| 5 | Server / network error | `server_error`, `network_error` |

## License

[MIT](LICENSE)
