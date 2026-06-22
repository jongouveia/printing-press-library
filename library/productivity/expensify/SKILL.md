---
name: pp-expensify
description: "File expenses and submit reports to Expensify in one line. Every command an agent should need, with a local cache so searches stay offline. Trigger phrases: `file an expense`, `submit my expense report`, `expense that`, `what did I expense this month`, `use expensify`, `run expensify`."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["expensify-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/productivity/expensify-pp-cli/cmd/expensify-pp-cli@latest","bins":["expensify-pp-cli"],"label":"Install via go install"}]}}'
---

# Expensify — Printing Press CLI

expensify-pp-cli turns the Expensify web app into a terminal. Log in once, and every filing/reviewing/submitting task that used to require clicking through forms becomes a single command. A local SQLite store gives you offline search, rollups, dupe detection, and missing-receipt alerts that no other Expensify tool has.

## Prerequisites: Install the CLI

This skill drives the `expensify-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer:
   ```bash
   npx -y @mvanhorn/printing-press-library install expensify --cli-only
   ```
2. Verify: `expensify-pp-cli --version`
3. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.3 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/productivity/expensify/cmd/expensify-pp-cli@latest
```

If `--version` reports "command not found" after install, the install step did not put the binary on `$PATH`. Do not proceed with skill commands until verification succeeds.

## When to Use This CLI

Pick this CLI when you want to file, review, or submit Expensify expenses and reports without leaving the terminal. It excels at one-liner expense filing from natural language, end-of-month report drafting from a date range, offline search across years of expenses, and orchestrating accounting-system exports. Agents can drive it through the standard --json output and typed exit codes, or through the built-in MCP bridge mode.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Filing at thought speed
- **`expense quick`** — File an expense with one line: amount, merchant, and category parsed from a short prompt — no forms, no web UI.

  _When a user tells an agent 'I just expensed dinner at Maya for $42.50,' the agent should file it in one call — not walk through six fields._

  ```bash
  expensify-pp-cli expense quick "Dinner at Maya $42.50" --agent
  ```
- **`report draft`** — Create a report and auto-attach every un-reported expense from a date range in a single command.

  _End-of-month submission turns from 45 clicks into one command._

  ```bash
  expensify-pp-cli report draft --since 2026-04-01 --title "April expenses" --policy 201C8BE1F19AACD2
  ```
- **`expense watch`** — Daemon mode — drop a receipt PNG/JPG/PDF into a folder, it auto-files as an expense.

  _A scanner app or Siri shortcut drops files into Dropbox; this files them automatically._

  ```bash
  expensify-pp-cli expense watch ~/Receipts --policy 201C8BE1F19AACD2
  ```
- **`expense from-line`** — Paste a raw bank or card CSV row and the CLI extracts date/merchant/amount/currency and files the expense.

  _Reconciling AmEx? Paste a row, file an expense. No copy-paste-copy-paste._

  ```bash
  expensify-pp-cli expense from-line "2026-04-18 DOORDASH*JOES $14.25" --category Meals
  ```
- **`expense quick`** — When filing an expense, suggests a category based on prior classifications for the same merchant.

  _Filing is faster when the category is already right 90% of the time._

  ```bash
  expensify-pp-cli expense quick "Uber $24"  # auto-suggests Transportation
  ```

### Local state that compounds
- **`damage`** — Single-glance summary: total expensed, pending, approved, paid for the current month (or a custom range).

  _Agents asked 'how much did I expense this month' get one answer in one call._

  ```bash
  expensify-pp-cli damage --month current --json
  ```
- **`expense search`** — FTS5 search over all your expenses by merchant, comment, category, or tag. Regex-friendly.

  _Agents asked 'did I expense that Starbucks last month' get an answer in one local query._

  ```bash
  expensify-pp-cli expense search "coffee" --since 2026-01-01 --json
  ```
- **`expense missing-receipts`** — Lists expenses without attached receipts so you can catch them before submitting a report.

  _Submit-report-and-get-bounced feels bad; surface missing receipts upfront._

  ```bash
  expensify-pp-cli expense missing-receipts --json
  ```
- **`expense rollup`** — Pivot-table expenses by category, tag, or merchant for any time range.

  _Build your own spending dashboard without burning API budget._

  ```bash
  expensify-pp-cli expense rollup --month 2026-04 --by category
  ```
- **`expense dupes`** — Finds expenses that look like duplicates by (merchant, amount, date±window).

  _Accidental double-file is a top AP pain point; surface it before submission._

  ```bash
  expensify-pp-cli expense dupes --window 3d --json
  ```

### Agent-native plumbing
- **`report submit --wait`** — Blocks until the report transitions past SUBMITTED, great for CI pipelines.

  _CI jobs that submit expense reports can wait for approval before moving on._

  ```bash
  expensify-pp-cli report submit 1587860702457827 --wait --timeout 1h
  ```
- **`undo`** — Revert the last create/edit/submit action from the local action log.

  _Agents sometimes file the wrong amount; one command to roll it back._

  ```bash
  expensify-pp-cli undo --dry-run
  ```
- **`mcp`** — Expose every subcommand as an MCP tool so Claude Desktop and other MCP clients can drive Expensify.

  _Claude Desktop gets a single-install Expensify integration that matches this CLI's coverage exactly._

  ```bash
  expensify-pp-cli mcp --port 7021
  ```

### Admin orchestration
- **`close`** — End-of-month orchestrator: list reports in range, export with GL template, download the file, mark-as-exported atomically.

  _Finance teams close a month in one command instead of a 40-click sequence._

  ```bash
  expensify-pp-cli close --month 2026-04 --template netsuite --label "Apr close"
  ```
- **`admin policy-diff`** — Compare local YAML policy config against live API to preview changes before applying.

  _Manage chart-of-accounts as code with dry-run safety._

  ```bash
  expensify-pp-cli admin policy diff 201C8BE1F19AACD2 categories.yaml
  ```

## Command Reference

**admin** — Integration Server: policy, employee, and rules admin

- `expensify-pp-cli admin cards_list` — List domain cards (Domain Cards Getter)
- `expensify-pp-cli admin cards_owners` — List card owners (Card Owner Data)
- `expensify-pp-cli admin employee_add` — Add an employee to a policy (Advanced Employee Updater)
- `expensify-pp-cli admin employee_remove` — Remove an employee from a policy
- `expensify-pp-cli admin employee_update` — Update an employee (Advanced Employee Updater)
- `expensify-pp-cli admin policy_get` — Get a policy's full config (Policy Getter)
- `expensify-pp-cli admin policy_list` — List all policies you admin (Policy List Getter)
- `expensify-pp-cli admin policy_new` — Create a new policy (Policy Creator)
- `expensify-pp-cli admin policy_set_categories` — Update categories for a policy from YAML
- `expensify-pp-cli admin policy_set_fields` — Update report fields for a policy
- `expensify-pp-cli admin policy_set_tags` — Update tags for a policy from YAML
- `expensify-pp-cli admin report_set_status` — Force a report status transition (Report Status Updater)
- `expensify-pp-cli admin rules_new` — Create an expense rule (Expense Rules Creator)
- `expensify-pp-cli admin rules_update` — Update an expense rule
- `expensify-pp-cli admin tag_approvers_set` — Set tag approvers (Tag Approvers Updater)

**category** — Workspace categories (for expense classification)

- `expensify-pp-cli category list` — List categories for a workspace

**expense** — Create, list, and manage personal expenses

- `expensify-pp-cli expense attach` — Attach or replace a receipt on an expense
- `expensify-pp-cli expense create` — Create a new expense
- `expensify-pp-cli expense delete` — Delete an expense
- `expensify-pp-cli expense edit` — Edit an existing expense
- `expensify-pp-cli expense get` — Get expense detail by transaction ID
- `expensify-pp-cli expense list` — List your expenses with filters

**export** — Integration Server: export reports to accounting systems (admin)

- `expensify-pp-cli export download` — Download a previously generated export file
- `expensify-pp-cli export run` — Export reports via Report Exporter (Integration Server)

**me** — Current user profile

- `expensify-pp-cli me get` — Get current user profile

**recon** — Integration Server: corporate card reconciliation (admin)

- `expensify-pp-cli recon export` — Export reconciliation data for a domain

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

- `expensify-pp-cli tag list` — List tags for a workspace

**workspace** — View workspaces (policies) you have access to

- `expensify-pp-cli workspace get` — Get workspace detail
- `expensify-pp-cli workspace list` — List workspaces accessible to your account

## Recipes

### File a quick expense

```bash
expensify-pp-cli expense quick "Lunch at Joe's $18.50"
```

One line, one call, one expense. Category auto-suggested from your history.

### Draft this month's report

```bash
expensify-pp-cli report draft --since 2026-04-01 --title "April" --policy <workspace-id>
```

Creates a report and attaches every un-reported April expense in a single command.

### Submit for approval and wait

```bash
expensify-pp-cli report submit <report-id> --wait --timeout 1h
```

Blocks until approval arrives — drop this in CI after a closing script.

### Find expenses with no receipt

```bash
expensify-pp-cli expense missing-receipts --json
```

Catches receipt gaps before they bounce your report.

### Close the month (admin)

```bash
expensify-pp-cli close --month 2026-04 --template netsuite --label "Apr close"
```

End-to-end: list approved reports, export with the NetSuite template, download the file, mark as exported.

## Auth Setup

Two ways to authenticate: (1) `expensify auth login` opens a browser, you log in, the CLI captures your session token — works immediately for all filing/submitting commands; (2) `expensify auth set-keys` stores your Integration Server partner credentials (get them at https://www.expensify.com/tools/integrations/) — required only for export/admin commands. Most users only need option 1.

Run `expensify-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select id,name` returns a subset of fields
- **Previewable** — `--dry-run` shows the request without sending
- **Cacheable** — GET responses cached for 5 minutes, bypass with `--no-cache`
- **Non-interactive** — never prompts, every input is a flag

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
<cli>-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
<cli>-pp-cli feedback --stdin < notes.txt
<cli>-pp-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.<cli>-pp-cli/feedback.jsonl`. They are never POSTed unless `<CLI>_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `<CLI>_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

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
<cli>-pp-cli profile save briefing --<flag> <value> ...
<cli>-pp-cli --profile briefing <command> ...
<cli>-pp-cli profile list [--json]
<cli>-pp-cli profile show briefing
<cli>-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Async Jobs

The long-running workflow in this CLI is `report submit`, which supports `--wait` to block until Expensify transitions the report past SUBMITTED, and `--timeout` (default 1h) to bound the wait.

| Flag | Purpose |
|------|---------|
| `--wait` | Block until the report leaves the SUBMITTED state (approved, reimbursed, or rejected). |
| `--timeout` | Maximum wait duration, e.g. `30m`, `1h`, `24h`. |

Use `report submit <id>` without `--wait` for fire-and-forget. Add `--wait --timeout 1h` when the caller needs the final status inline.

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
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/productivity/expensify-pp-cli/cmd/expensify-pp-cli@latest
   ```
3. Verify: `expensify-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/productivity/expensify-pp-cli/cmd/expensify-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add expensify-pp-mcp -- expensify-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which expensify-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   expensify-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `expensify-pp-cli <command> --help`.
