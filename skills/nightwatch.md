# Nightwatch CLI

CLI for the Nightwatch.io SEO API. All output is valid JSON to stdout. Errors go to stderr as JSON.

Install: `go install github.com/nightwatch-seo/nightwatch-cli@latest`

## Setup

Authentication (checked in order):
1. `NIGHTWATCH_API_KEY` environment variable
2. `~/.config/nightwatch/config.yaml`

```bash
# Option 1: env var
export NIGHTWATCH_API_KEY=your-token-here

# Option 2: config file
nightwatch config set api-key YOUR_TOKEN
```

## Global Flags

All commands accept these flags:
- `--dry-run` — Preview the HTTP request as JSON without calling the API
- `--base-url URL` — Override API base URL (default: `https://api.nightwatch.io/api/v1`)
- `--timeout DURATION` — Per-request timeout (default: `30s`)
- `--no-retry` — Disable automatic retry on HTTP 429

## Response Format

- Success: `{"data": ..., "meta": {...}}`
- Error: `{"error": "message", "code": N, "error_type": "type"}`
- Exit codes: 0=success, 1=client error, 2=auth error, 3=rate-limit, 5=server error

---

## Commands Reference

### Groups — Organize tracked URLs

```bash
# List all groups
nightwatch groups list

# Get a group by ID
nightwatch groups get 123

# Create a group
nightwatch groups create --name "My SEO Projects"

# Rename a group
nightwatch groups update 123 --name "Renamed Group"

# Delete a group
nightwatch groups delete 123
```

### URLs — Tracked websites

```bash
# List all URLs (optionally filter by group)
nightwatch urls list
nightwatch urls list --group-id 42

# Get a URL by ID
nightwatch urls get 456

# Create a tracked URL
nightwatch urls create \
  --url "https://example.com" \
  --group-id 42 \
  --country-code us \
  --language-code en

# Create with optional settings
nightwatch urls create \
  --url "https://shop.example.com" \
  --group-id 42 \
  --country-code gb \
  --language-code en \
  --custom-name "UK Shop" \
  --match-subdomains \
  --match-nested-urls

# Update URL settings
nightwatch urls update 456 --custom-name "Updated Name" --country-code gb

# Delete a URL (removes all keywords, views, history)
nightwatch urls delete 456
```

**urls create** required flags: `--url`, `--group-id`, `--country-code`, `--language-code`

**urls update** optional flags: `--url`, `--custom-name`, `--country-code`, `--language-code`, `--group-id`, `--places-match`, `--match-subdomains`, `--match-nested-urls`, `--include-local-pack`, `--include-places-image`, `--include-featured-snippet`, `--include-knowledge-panel`, `--site-audit-interval`

### Keywords — Tracked search terms

```bash
# List keywords for a URL
nightwatch keywords list --url-id 123
nightwatch keywords list --url-id 123 --page 2 --limit 50
nightwatch keywords list --url-id 123 --search "seo" --sort position --direction asc
nightwatch keywords list --url-id 123 --view-id 789

# Add keywords (comma-separated)
nightwatch keywords add \
  --url-id 123 \
  --keywords "seo tools,keyword tracking" \
  --google-gl us \
  --google-hl en \
  --search-engine google

# Add with optional settings
nightwatch keywords add \
  --url-id 123 \
  --keywords "local seo" \
  --google-gl us \
  --google-hl en \
  --search-engine google \
  --mobile \
  --tags "local,priority"

# Remove keywords by ID
nightwatch keywords remove --url-id 123 --keyword-ids 456,789,101

# Update keyword tags
nightwatch keywords update --url-id 123 --keyword-id 456 --tags "priority,brand"
```

**keywords list** optional flags: `--view-id`, `--page`, `--limit`, `--search`, `--sort`, `--direction` (asc/desc)

**keywords add** required flags: `--url-id`, `--keywords`, `--google-gl`, `--google-hl`, `--search-engine`
Optional: `--mobile`, `--desktop`, `--tags`, `--adwords-location-id`

### Views — Dynamic keyword filters

```bash
# List views (optionally filter by URL or group)
nightwatch views list
nightwatch views list --url-id 123
nightwatch views list --group-id 42 --without-counts

# Get a view by ID
nightwatch views get 789

# Create a view (scoped to URL or group)
nightwatch views create --name "Top 10 Keywords" --url-id 123
nightwatch views create --name "Brand Keywords" --group-id 42

# Rename a view
nightwatch views update 789 --name "Renamed View"

# Add filter rules to a view (JSON array)
nightwatch views add-filter 789 \
  --filters '[{"field":"query","condition":"contains","value":"seo"}]'
nightwatch views add-filter 789 \
  --filters '[{"field":"position","condition":"less_than","value":"10"}]'

# Delete a view
nightwatch views delete 789
```

**views add-filter** filter object fields: `field` (query, position, tags, ...), `condition` (contains, equals, greater_than, less_than, ...), `value`

### Competitors — Track rival URLs

```bash
# List competitors for a URL
nightwatch competitors list --url-id 123

# Add a competitor
nightwatch competitors add --url-id 123 --url "https://competitor.com"
nightwatch competitors add --url-id 123 --url "https://competitor.com" \
  --custom-name "Main Competitor" --match-subdomains

# Remove a competitor
nightwatch competitors remove --url-id 123 --competitor-id 456
```

**competitors add** optional flags: `--custom-name`, `--match-subdomains`, `--match-nested-urls`, `--places-match`, `--places-image-title`

### Series — Historical ranking data

```bash
# Get ranking history for URLs
nightwatch series get \
  --date-from 2025-01-01 \
  --date-to 2025-03-01 \
  --url-ids 123

# Get history for specific keywords with competitor data
nightwatch series get \
  --date-from 2025-01-01 \
  --date-to 2025-03-01 \
  --keyword-ids 456,789 \
  --with-competitors
```

Required: `--date-from`, `--date-to`, and at least one of: `--url-ids`, `--keyword-ids`, `--view-ids`, `--group-ids`
Optional: `--with-competitors`

### Stats — Keyword statistics

```bash
# Stats for a URL
nightwatch stats get --start-date 2025-01-01 --end-date 2025-03-01 --url-id 123

# Stats for a view with organic metrics
nightwatch stats get --start-date 2025-01-01 --end-date 2025-03-01 --view-id 789 --include-organic
```

Required: `--start-date`, `--end-date`, and one of: `--url-id` or `--view-id`
Optional: `--include-organic`

### SERPs — Search results page data

```bash
# Get SERP data by keyword result ID
nightwatch serps get 12345
```

### Config — CLI settings

```bash
# Save API key
nightwatch config set api-key YOUR_TOKEN

# Override base URL
nightwatch config set base-url https://api.example.com/v1

# Show resolved config (API key is masked)
nightwatch config show
```

---

## Common Workflows

### Set up a new site for tracking
```bash
# 1. Create a group
nightwatch groups create --name "Client Sites"
# 2. Create a tracked URL (use the group ID from step 1)
nightwatch urls create --url "https://example.com" --group-id GROUP_ID --country-code us --language-code en
# 3. Add keywords (use the URL ID from step 2)
nightwatch keywords add --url-id URL_ID --keywords "seo,keyword research,rank tracking" --google-gl us --google-hl en --search-engine google
```

### Check ranking trends
```bash
# List keywords and their current positions
nightwatch keywords list --url-id 123 --sort position --direction asc --limit 50
# Get historical data for a date range
nightwatch series get --date-from 2025-01-01 --date-to 2025-03-01 --url-ids 123
```

### Monitor competitors
```bash
# Add competitors
nightwatch competitors add --url-id 123 --url "https://rival.com"
# Get ranking series with competitor data
nightwatch series get --date-from 2025-01-01 --date-to 2025-03-01 --url-ids 123 --with-competitors
```

### Create a filtered keyword view
```bash
# Create view
nightwatch views create --name "Top 10" --url-id 123
# Add filter: only keywords in top 10
nightwatch views add-filter VIEW_ID --filters '[{"field":"position","condition":"less_than","value":"11"}]'
# List keywords through the view
nightwatch keywords list --url-id 123 --view-id VIEW_ID
```

## Tips

- Use `--dry-run` on any command to preview the HTTP request without calling the API.
- Pipe output to `jq` for filtering: `nightwatch keywords list --url-id 123 | jq '.data[].query'`
- All IDs are numeric. Get them from the `list` or `create` commands first.
- Dates use `YYYY-MM-DD` format.
- Country/language codes are two-letter ISO codes (e.g. `us`, `gb`, `en`, `de`).
- Search engines for keyword tracking: `google`, `bing`, `youtube`.
