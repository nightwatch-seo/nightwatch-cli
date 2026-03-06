# Nightwatch CLI — Engineering Standards

> Agent-first CLI for the Nightwatch.io SEO API. Go + cobra.

---

## Project Overview

- **Language:** Go
- **CLI framework:** [cobra](https://github.com/spf13/cobra)
- **Config format:** YAML (`gopkg.in/yaml.v3`)
- **API docs:** https://docs.nightwatch.io/
- **Architecture:** Follows the exact patterns of shovels-cli

## Architecture

```
cmd/            cobra command tree (one file per resource)
  root.go       root command, global flags, helpers (dry-run, client, error handling)
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

Unlike shovels-cli (GET only), nightwatch-cli uses GET, POST, PUT, and DELETE.
All POST/PUT requests send JSON bodies with `Content-Type: application/json`.

## Build

```bash
go build -o nightwatch .
./nightwatch --help
```

## Conventions

- One cobra command file per resource in `cmd/`
- All HTTP calls go through `internal/client/` — commands never call `net/http` directly
- Flag names use kebab-case (e.g. `--url-id`, `--group-id`)
- API parameters use snake_case (e.g. `url_group_id`, `country_code`)
- Wrap API responses: `{"data": ..., "meta": {...}}`
- Errors: `{"error": "msg", "code": N, "error_type": "type"}`
