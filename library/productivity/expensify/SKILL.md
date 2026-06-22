---
name: pp-expensify
description: "Printing Press CLI for Expensify. File expenses and submit reports to Expensify from the command line"
author: "Matt Van Horn"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - expensify-pp-cli
    install:
      - kind: go
        bins: [expensify-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/productivity/expensify/cmd/expensify-pp-cli
---

# Expensify — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `expensify-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install expensify --cli-only
   ```
2. Verify: `expensify-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.4 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/productivity/expensify/cmd/expensify-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

File expenses and submit reports to Expensify from the command line

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

## Command Reference

**admin** — Integration Server: policy, employee, and rules admin

- `expensify-pp-cli admin cards-list` — List domain cards (Domain Cards Getter)
- `expensify-pp-cli admin cards-owners` — List card owners (Card Owner Data)
- `expensify-pp-cli admin employee-add` — Add an employee to a policy (Advanced Employee Updater)
- `expensify-pp-cli admin employee-remove` — Remove an employee from a policy
- `expensify-pp-cli admin employee-update` — Update an employee (Advanced Employee Updater)
- `expensify-pp-cli admin policy-get` — Get a policy's full config (Policy Getter)
- `expensify-pp-cli admin policy-list` — List all policies you admin (Policy List Getter)
- `expensify-pp-cli admin policy-new` — Create a new policy (Policy Creator)
- `expensify-pp-cli admin policy-set-categories` — Update categories for a policy from YAML
- `expensify-pp-cli admin policy-set-fields` — Update report fields for a policy
- `expensify-pp-cli admin policy-set-tags` — Update tags for a policy from YAML
- `expensify-pp-cli admin report-set-status` — Force a report status transition (Report Status Updater)
- `expensify-pp-cli admin rules-new` — Create an expense rule (Expense Rules Creator)
- `expensify-pp-cli admin rules-update` — Update an expense rule
- `expensify-pp-cli admin tag-approvers-set` — Set tag approvers (Tag Approvers Updater)

**category** — Workspace categories (for expense classification)

- `expensify-pp-cli category` — List categories for a workspace

**expense** — Create, list, and manage personal expenses

- `expensify-pp-cli expense attach` — Attach or replace a receipt on an expense
- `expensify-pp-cli expense create` — Create a new expense
- `expensify-pp-cli expense delete` — Delete an expense
- `expensify-pp-cli expense edit` — Edit an existing expense
- `expensify-pp-cli expense get` — Get expense detail by transaction ID
- `expensify-pp-cli expense list` — List your expenses with filters

**export_resource** — Integration Server: export reports to accounting systems (admin)

- `expensify-pp-cli export-resource download` — Download a previously generated export file
- `expensify-pp-cli export-resource run` — Export reports via Report Exporter (Integration Server)

**me** — Current user profile

- `expensify-pp-cli me` — Get current user profile

**recon** — Integration Server: corporate card reconciliation (admin)

- `expensify-pp-cli recon` — Export reconciliation data for a domain

**report** — Create, manage, and submit expense reports

- `expensify-pp-cli report add` — Add expenses to a report
- `expensify-pp-cli report approve` — Approve a report (manager action)
- `expensify-pp-cli report comment` — Add a comment to a report thread
- `expensify-pp-cli report create` — Create a new report
- `expensify-pp-cli report delete` — Delete a draft report
- `expensify-pp-cli report get` — Get report detail
- `expensify-pp-cli report list` — List your reports
- `expensify-pp-cli report pay` — Mark a report as reimbursed
- `expensify-pp-cli report reopen` — Reopen a submitted report back to draft
- `expensify-pp-cli report submit` — Submit a report for approval

**tag** — Workspace tags (multi-level, for expense classification)

- `expensify-pp-cli tag` — List tags for a workspace

**workspace** — View workspaces (policies) you have access to

- `expensify-pp-cli workspace get` — Get workspace detail
- `expensify-pp-cli workspace list` — List workspaces accessible to your account


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
expensify-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Auth Setup
Run `expensify-pp-cli auth setup` to print the URL and steps for getting a key (add `--launch` to open the URL). Then set:

```bash
export EXPENSIFY_AUTH_TOKEN="<your-key>"
```

Or persist it in `~/.config/expensify-pp-cli/config.toml`.

Run `expensify-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  expensify-pp-cli category --policy-id 550e8400-e29b-41d4-a716-446655440000 --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Non-interactive** — never prompts, every input is a flag
- **Explicit retries** — use `--idempotent` only when an already-existing create should count as success

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
expensify-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
expensify-pp-cli feedback --stdin < notes.txt
expensify-pp-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.local/share/expensify-pp-cli/feedback.jsonl`. They are never POSTed unless `EXPENSIFY_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `EXPENSIFY_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
expensify-pp-cli profile save briefing --json
expensify-pp-cli --profile briefing category --policy-id 550e8400-e29b-41d4-a716-446655440000
expensify-pp-cli profile list --json
expensify-pp-cli profile show briefing
expensify-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `expensify-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/productivity/expensify/cmd/expensify-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add expensify-pp-mcp -- expensify-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which expensify-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   expensify-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `expensify-pp-cli <command> --help`.
