# Expensify CLI

File expenses and submit reports to Expensify from the command line

Learn more at [Expensify](https://www.expensify.com/).

Created by [@mvanhorn](https://github.com/mvanhorn) (Matt Van Horn).

## Install

The recommended path installs both the `expensify-pp-cli` binary and the `pp-expensify` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install expensify
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install expensify --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install expensify --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install expensify --agent claude-code
npx -y @mvanhorn/printing-press-library install expensify --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.4 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/productivity/expensify/cmd/expensify-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/expensify-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install expensify --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-expensify --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-expensify --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install expensify --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/expensify-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.
3. Fill in `EXPENSIFY_AUTH_TOKEN` when Claude Desktop prompts you.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


```bash
go install github.com/mvanhorn/printing-press-library/library/productivity/expensify/cmd/expensify-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "expensify": {
      "command": "expensify-pp-mcp",
      "env": {
        "EXPENSIFY_AUTH_TOKEN": "<your-key>"
      }
    }
  }
}
```

</details>

## Quick Start

### 1. Install

See [Install](#install) above.

### 2. Set Up Credentials

Get your API key from your API provider's developer portal. The key typically looks like a long alphanumeric string.

```bash
export EXPENSIFY_AUTH_TOKEN="<paste-your-key>"
```

You can also persist this in your config file at `~/.config/expensify-pp-cli/config.toml`.

### 3. Verify Setup

```bash
expensify-pp-cli doctor
```

This checks your configuration and credentials.

### 4. Try Your First Command

```bash
expensify-pp-cli category --policy-id 550e8400-e29b-41d4-a716-446655440000
```

## Usage

Run `expensify-pp-cli --help` for the full command reference and flag list.

## Commands

### admin

Integration Server: policy, employee, and rules admin

- **`expensify-pp-cli admin cards-list`** - List domain cards (Domain Cards Getter)
- **`expensify-pp-cli admin cards-owners`** - List card owners (Card Owner Data)
- **`expensify-pp-cli admin employee-add`** - Add an employee to a policy (Advanced Employee Updater)
- **`expensify-pp-cli admin employee-remove`** - Remove an employee from a policy
- **`expensify-pp-cli admin employee-update`** - Update an employee (Advanced Employee Updater)
- **`expensify-pp-cli admin policy-get`** - Get a policy's full config (Policy Getter)
- **`expensify-pp-cli admin policy-list`** - List all policies you admin (Policy List Getter)
- **`expensify-pp-cli admin policy-new`** - Create a new policy (Policy Creator)
- **`expensify-pp-cli admin policy-set-categories`** - Update categories for a policy from YAML
- **`expensify-pp-cli admin policy-set-fields`** - Update report fields for a policy
- **`expensify-pp-cli admin policy-set-tags`** - Update tags for a policy from YAML
- **`expensify-pp-cli admin report-set-status`** - Force a report status transition (Report Status Updater)
- **`expensify-pp-cli admin rules-new`** - Create an expense rule (Expense Rules Creator)
- **`expensify-pp-cli admin rules-update`** - Update an expense rule
- **`expensify-pp-cli admin tag-approvers-set`** - Set tag approvers (Tag Approvers Updater)

### category

Workspace categories (for expense classification)

- **`expensify-pp-cli category`** - List categories for a workspace

### expense

Create, list, and manage personal expenses

- **`expensify-pp-cli expense attach`** - Attach or replace a receipt on an expense
- **`expensify-pp-cli expense create`** - Create a new expense
- **`expensify-pp-cli expense delete`** - Delete an expense
- **`expensify-pp-cli expense edit`** - Edit an existing expense
- **`expensify-pp-cli expense get`** - Get expense detail by transaction ID
- **`expensify-pp-cli expense list`** - List your expenses with filters

### export_resource

Integration Server: export reports to accounting systems (admin)

- **`expensify-pp-cli export-resource download`** - Download a previously generated export file
- **`expensify-pp-cli export-resource run`** - Export reports via Report Exporter (Integration Server)

### me

Current user profile

- **`expensify-pp-cli me`** - Get current user profile

### recon

Integration Server: corporate card reconciliation (admin)

- **`expensify-pp-cli recon`** - Export reconciliation data for a domain

### report

Create, manage, and submit expense reports

- **`expensify-pp-cli report add`** - Add expenses to a report
- **`expensify-pp-cli report approve`** - Approve a report (manager action)
- **`expensify-pp-cli report comment`** - Add a comment to a report thread
- **`expensify-pp-cli report create`** - Create a new report
- **`expensify-pp-cli report delete`** - Delete a draft report
- **`expensify-pp-cli report get`** - Get report detail
- **`expensify-pp-cli report list`** - List your reports
- **`expensify-pp-cli report pay`** - Mark a report as reimbursed
- **`expensify-pp-cli report reopen`** - Reopen a submitted report back to draft
- **`expensify-pp-cli report submit`** - Submit a report for approval

### tag

Workspace tags (multi-level, for expense classification)

- **`expensify-pp-cli tag`** - List tags for a workspace

### workspace

View workspaces (policies) you have access to

- **`expensify-pp-cli workspace get`** - Get workspace detail
- **`expensify-pp-cli workspace list`** - List workspaces accessible to your account


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
expensify-pp-cli category --policy-id 550e8400-e29b-41d4-a716-446655440000

# JSON for scripting and agents
expensify-pp-cli category --policy-id 550e8400-e29b-41d4-a716-446655440000 --json

# Filter to specific fields
expensify-pp-cli category --policy-id 550e8400-e29b-41d4-a716-446655440000 --json --select id,name,status

# Dry run — show the request without sending
expensify-pp-cli category --policy-id 550e8400-e29b-41d4-a716-446655440000 --dry-run

# Agent mode — JSON + compact + no prompts in one flag
expensify-pp-cli category --policy-id 550e8400-e29b-41d4-a716-446655440000 --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Explicit retries** - add `--idempotent` to create retries when a no-op success is acceptable
- **Confirmable** - `--yes` for explicit confirmation of destructive actions
- **Piped input** - write commands can accept structured input when their help lists `--stdin`
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
expensify-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Config file: `~/.config/expensify-pp-cli/config.toml`

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `EXPENSIFY_AUTH_TOKEN` | per_call | Yes | Set to your API credential. |
| `EXPENSIFY_PARTNER_USER_ID` | per_call | Yes | Set to your API credential. |
| `EXPENSIFY_PARTNER_USER_SECRET` | per_call | Yes | Set to your API credential. |

### agentcookie (optional)

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `expensify-pp-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `expensify-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $EXPENSIFY_AUTH_TOKEN`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
