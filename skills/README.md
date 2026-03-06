# Nightwatch CLI — AI Skills

These files teach AI coding assistants how to use the Nightwatch CLI. Pick the one that matches your tool.

## Files

| File | Platform | How to install |
|------|----------|----------------|
| `nightwatch.md` | Any AI / LLM | Paste into your system prompt or project context |
| `claude-code.md` | Claude Code | Copy `nightwatch.md` to `.claude/skills/nightwatch/SKILL.md` in your project |
| `cursor.md` | Cursor | Copy `nightwatch.md` into `.cursor/rules/nightwatch.mdc` with the frontmatter shown |

## Quick start

### Claude Code

```bash
# From your project root:
mkdir -p .claude/skills/nightwatch
cp path/to/nightwatch-cli/skills/nightwatch.md .claude/skills/nightwatch/SKILL.md
```

Then add the YAML frontmatter at the top of the file:

```yaml
---
name: nightwatch
description: |
  Use the Nightwatch CLI to manage SEO keyword tracking, URL monitoring,
  competitor analysis, ranking history, and SERP data via the Nightwatch.io API.
  Trigger when the user asks about SEO tracking, keyword rankings, search engine
  positions, SERP data, competitor monitoring, or anything related to Nightwatch.
---
```

### Cursor

```bash
mkdir -p .cursor/rules
cp path/to/nightwatch-cli/skills/nightwatch.md .cursor/rules/nightwatch.mdc
```

Then add the MDC frontmatter at the top of the file:

```mdc
---
description: Use the Nightwatch CLI for SEO keyword tracking, URL monitoring, competitor analysis, ranking history, and SERP data.
globs:
alwaysApply: false
---
```

### Windsurf / Copilot / Other

Paste the contents of `nightwatch.md` into your project's AI rules file (`.windsurfrules`, `.github/copilot-instructions.md`, etc).
