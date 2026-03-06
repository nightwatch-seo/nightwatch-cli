# Nightwatch CLI — Engineering Standards

> Agent-first CLI for the Nightwatch.io SEO API. Go + cobra.

---

## Project Overview

- **Language:** Go
- **CLI framework:** [cobra](https://github.com/spf13/cobra)
- **Config format:** YAML (`gopkg.in/yaml.v3`)
- **API docs:** https://docs.nightwatch.io/
- **Architecture:** cobra command tree + internal packages for client, config, output

## Architecture

```
cmd/            cobra command tree (one file per resource)
  root.go       root command, global flags, shared helpers
  groups.go     URL group CRUD
  urls.go       tracked URL CRUD
  keywords.go   keyword list/add/remove/update
  views.go      keyword view CRUD + add-filter
  competitors.go  competitor list/add/remove
  series.go     historical data series
  stats.go      keyword statistics
  serps.go      SERP data retrieval
  config.go     config set/show
  version.go    version output
internal/
  client/       HTTP client (Get/Post/Put/Delete + retry)
  config/       config file + env var resolution
  output/       JSON output formatting, error rendering
```

## Design Principles

- **JSON-only output.** No tables, no colors. Every response is valid JSON to stdout.
- **Errors to stderr.** Structured JSON errors go to stderr so stdout is always parseable.
- **Never interactive.** No prompts, no spinners, no progress bars. Fail loudly with clear messages.
- **Meaningful exit codes:** 0=success, 1=client error, 2=auth error, 3=rate-limit, 5=transient/server error.
- **`--help` text is for LLMs.** Write descriptions as if an AI agent is reading them — specific, example-rich, no jargon.
- **`--dry-run` previews requests.** Shows the resolved HTTP method, URL, and body as JSON without calling the API.

## Auth

- Header: `Authorization: <token>` (not X-API-Key)
- Env var: `NIGHTWATCH_API_KEY`
- Config file: `~/.config/nightwatch/config.yaml`
- Precedence: env var > config file

## API Base URL

`https://api.nightwatch.io/api/v1`

## HTTP Methods

The HTTP client supports GET, POST, PUT, and DELETE.
All POST/PUT requests send JSON bodies with `Content-Type: application/json`.

## Multi-Host Support

The HTTP client supports absolute URLs for endpoints on different hosts.
Pass a full URL (starting with `https://`) instead of a relative path:

```go
// Relative path — prepends baseURL automatically
cl.Get(ctx, "/url_groups", nil)

// Absolute URL — used as-is, skips baseURL
cl.Get(ctx, "https://other-api.nightwatch.io/v2/backlinks", nil)
```

The `resolveEndpoint()` helper in `cmd/root.go` applies the same logic for `--dry-run` output.

## Build

```bash
go build -o nightwatch .
./nightwatch --help
```

## Shared Helpers (cmd/root.go)

| Helper | Purpose |
|--------|---------|
| `newClientFromFlags(cmd)` | Creates `client.Client` from resolved config + flags |
| `handleAPIError(err)` | Translates `*client.APIError` to JSON stderr + exitError |
| `printResponse(cmd, body, fn)` | Parses response body and prints via `printAsSingle` or `printAsList` |
| `printDryRunGet/Post/Put/Delete` | Prints resolved request as JSON for `--dry-run` |
| `validateDate(flag, value)` | Validates YYYY-MM-DD format |
| `isDryRun(cmd)` | Checks `--dry-run` flag |
| `resolveEndpoint(endpoint)` | Prepends baseURL unless endpoint is already absolute |

## Conventions

- One cobra command file per resource in `cmd/`
- All HTTP calls go through `internal/client/` — commands never call `net/http` directly
- Flag names use kebab-case (e.g. `--url-id`, `--group-id`)
- API parameters use snake_case (e.g. `url_group_id`, `country_code`)
- Wrap API responses: `{"data": ..., "meta": {...}}`
- Errors: `{"error": "msg", "code": N, "error_type": "type"}`

## Not Implemented (v1)

These features are intentionally excluded from v1:
- `--schema` flag (offline response schema introspection)
- Auto-update mechanism
- `--limit all` pagination abstraction
- Credit tracking headers
