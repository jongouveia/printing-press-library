---
title: "fix: Repair expensify-pp-cli auth capture and expense-create API correctness"
type: fix
date: 2026-06-19
status: ready
---

# fix: Repair expensify-pp-cli auth capture and expense-create API correctness

**Target repo:** the `expensify-pp-cli` source. Canonical home is `mvanhorn/printing-press-library` (published via printing-press-publish/amend). The local working copy used by `go install` lives at `~/printing-press/library/expensify` and is **not** itself a git repo — confirm the actual PR target / branch before landing (see Open Questions). All paths below are relative to the CLI module root (the dir containing `go.mod`, `internal/`, `cmd/`).

---

## Summary

Two independent defects make the CLI unusable end-to-end, both confirmed live on 2026-06-18/19:

1. **Auth can't capture a session token.** `auth login` drives a throwaway `agent-browser` profile on `new.expensify.com`, polls for the `authToken` cookie for 2 minutes, and never sees one — because the user logs into their own Chrome on **classic** `www.expensify.com`, not the CLI's daemon profile. A second run dead-ends on "daemon already running". The paste fallback throws `reading stdin: EOF` without a TTY, and `--no-input` hard-fails outright. Net: every authenticated call returns `jsonCode 407 "session expired"`.

2. **Expense create silently drops the date.** All three create paths (`expense create`, `expense quick`, `expense from-line`) POST to `/RequestMoney` and send the date under the field name `date`, but the API's canonical field is `created` (confirmed by `internal/cli/sync.go`, which reads expenses back via `created`). The result: dates are ignored and every expense lands as "today" — exactly the symptom worked around by hand when filing 112 expenses. Amount-in-cents and `tag`=department are already correct.

This plan (a) makes token acquisition reliable in interactive, piped, and existing-browser contexts; and (b) fixes the date field, verifies the create command/param mapping against the live API, and adds a true bulk-create path so a long list of new expenses files in one request instead of one-at-a-time.

---

## Problem Frame

The CLI talks to the OldDot command dispatcher correctly at the transport layer (`internal/client/client.go` — form-encoded POST to `www.expensify.com/api/<Command>` with `authToken` in the body, no Authorization header, no csrfToken needed because token auth is used). The failures are at two specific seams: **token acquisition** (auth) and **create-command field/command correctness** (API). Both were validated against the live classic Expensify API: a form-encoded create with `merchant, amount (cents), created (YYYY-MM-DD), currency, category, tag(=department), reportID, reimbursable, comment` works, and the bulk command `Expense_Create` accepts a `transactionList` array (the client already JSON-encodes nested values into a single form field, so an array body field works without transport changes).

---

## Requirements

- R1. `auth login` reliably persists a working session `authToken` for classic `www.expensify.com`, or fails with an actionable message — never a silent 2-minute timeout that ends in an EOF.
- R2. Token entry works non-interactively: piped stdin, an explicit flag, and `--no-input` must all be able to set a token without a TTY.
- R3. Re-running `auth login` must not dead-end on a stale `agent-browser` daemon.
- R4. The CLI can acquire the token from the user's **existing** signed-in Chrome session, not only a throwaway profile.
- R5. Created expenses carry the date the user specified (`created`), across `expense create`, `quick`, `from-line`, and `edit`.
- R6. The create command and param mapping (command name, cents, `tag`=department, category, reimbursable) are verified against the live API and corrected where wrong.
- R7. A list of new expenses can be created in a single bulk request via `Expense_Create` + `transactionList`.
- R8. `doctor` and help/skill text reflect the corrected flows so a fresh user can self-serve.

---

## Key Technical Decisions

- **KTD1 — Token auth only; no csrfToken in the CLI.** `client.go` authenticates every command with `authToken` in the form body. The csrfToken needed during the manual browser reverse-engineering is a browser-session artifact and is **not** required here. The whole auth problem reduces to "get a valid `authToken` into config". This keeps the API-side fix small.
- **KTD2 — Fix `date`→`created` rather than inventing new fields.** `sync.go` already treats `created` as canonical on read; the create path is simply using the wrong write key. Rename the body key in all create/edit paths; keep the user-facing `--date` flag name.
- **KTD3 — Verify command before assuming.** The CLI uses `/RequestMoney`; the live-verified create used the manual web form and `Expense_Create`. `RequestMoney` may or may not honor `created`/`tag`/`category` for a self-expense. Execution must confirm `/RequestMoney` with `created` produces a correctly-dated, correctly-departmented expense; if it does not, route single-create through `Expense_Create` with a one-element `transactionList` (same code path as bulk). This is a verify-then-pick, not a blind switch.
- **KTD4 — Bulk via `Expense_Create` + `transactionList`.** The transport already JSON-encodes a slice body field into one form value (`buildNewExpensifyRequest`), so bulk needs no client changes — only a command that assembles the array and posts once. One request, one token, avoids the per-row fragility seen with sequential writes.
- **KTD5 — Existing-Chrome capture is best-effort with a guaranteed manual fallback.** Reading the live Chrome `authToken` cookie is non-trivial on macOS: Chrome 127+ uses App-Bound Encryption on the cookie store, so a raw SQLite read returns ciphertext. Prefer a Chrome DevTools Protocol attach (read the cookie from a Chrome already running with remote debugging) or reuse existing cookie-extraction tooling; never make the robust manual path (R2) depend on it. If decryption/attach isn't available, the CLI must degrade to a clear "copy authToken from DevTools and pipe it in" instruction, not an error.
- **KTD6 — Idempotency is the caller's responsibility for bulk.** `Expense_Create` is not idempotent; a re-run duplicates. Bulk-create must support `--dry-run` preview and report exactly how many were created, so the user can verify a total rather than re-fire blindly.

---

## Implementation Units

### U1. Bulletproof non-interactive token entry

**Goal:** Setting a token never requires a TTY. (R2)
**Requirements:** R2
**Dependencies:** none
**Files:** `internal/cli/auth.go`, `internal/cli/auth_test.go`
**Approach:** In `newAuthSetTokenCmd`, accept the token from (a) the positional arg, (b) `-` meaning read stdin, or (c) piped stdin when no arg is given and stdin is not a TTY (use `github.com/mattn/go-isatty`, already a dependency). In `fallbackPromptForToken`, when `--no-input` is set **and** stdin has piped data, read and save it instead of erroring; only error when there is genuinely no token source. Keep `auth login --token` working unchanged. Trim whitespace/newlines from all paths.
**Patterns to follow:** `import.go` already reads stdin via `inputFile == "-"`; mirror that convention. `isTerminal()` usage already exists in `expense_create.go`.
**Test scenarios:**
- `auth set-token <token>` positional still saves (happy path).
- `echo <token> | auth set-token -` saves the piped value, trimmed.
- `echo <token> | auth set-token` (no arg, non-TTY stdin) saves the piped value.
- `auth set-token` with empty stdin and no arg → clear "no token provided" error, non-zero exit.
- `auth login --no-input` with a token piped on stdin → saves it (no longer hard-fails).
- `auth login --no-input` with no token source → actionable error naming `--token`/stdin, exit code 4 (auth).

### U2. Harden the browser login flow

**Goal:** `auth login` targets the right site and survives re-runs. (R1, R3)
**Requirements:** R1, R3
**Dependencies:** none
**Files:** `internal/cli/auth.go`, `internal/cli/auth_test.go`
**Approach:** Before opening, run `agent-browser --session <name> close` (best-effort, ignore "not running") so a stale daemon can't block `--headed open`. Point the opened URL and the cookie-domain expectation at classic `www.expensify.com` (the host the API and the user's data live on); keep `new.expensify.com` acceptable as a cookie source since both share the `expensify.com` domain. On poll timeout, the message must route to the U1 manual path with the exact DevTools steps rather than dead-ending. Make the 2-minute timeout and poll interval flags (`--timeout`, default 2m) so a slow login isn't guillotined.
**Patterns to follow:** existing `exec.Command("agent-browser", ...)` calls in `auth.go`.
**Test scenarios:**
- Stale daemon present → `close` is invoked before `open`; `open` succeeds (mock `agent-browser` via PATH shim).
- `agent-browser` absent → falls back to U1 manual path, no panic.
- Poll finds `authToken` on `www.expensify.com` domain → saved.
- Poll finds `authToken` only on `new.expensify.com` domain → still saved (domain match is suffix-based).
- Timeout → message names the manual `--token`/stdin path; exit is the fallback path, not a bare error.

### U3. Capture the token from an existing Chrome session

**Goal:** Pull `authToken` from the user's already-signed-in Chrome without a throwaway profile. (R4)
**Requirements:** R4
**Dependencies:** U1 (manual fallback must exist first)
**Files:** `internal/cli/auth.go`, `internal/auth/chrome.go` (new), `internal/auth/chrome_test.go` (new)
**Approach:** Add `auth login --from-chrome` (and/or `auth from-browser`). Preferred mechanism: attach over Chrome DevTools Protocol to a Chrome started with `--remote-debugging-port` and read the `authToken` cookie for `www.expensify.com` via `Network.getCookies`/`Storage.getCookies`. If no debuggable Chrome is found, emit a one-line instruction to relaunch Chrome with remote debugging, or fall back to U1. Treat raw cookie-store decryption (App-Bound Encryption, Chrome 127+) as a separate, optional strategy behind the same command — do not block U3 on it (see KTD5, Risks). Never print the token; save directly to config and confirm with char count only.
**Execution note:** Spike the CDP-attach read against a live Chrome before committing to the SQLite-decrypt path; pick whichever reliably returns the cookie on this machine.
**Patterns to follow:** the project's prior browser-session token work (CDP attach + Chrome cookie handling) is the reference for mechanism choice.
**Test scenarios:**
- Mock CDP endpoint returns a cookie set including `authToken` for `www.expensify.com` → saved.
- Mock CDP returns no `authToken` → clear "not signed in / relaunch with remote debugging" message, falls back to U1, no token written.
- No debuggable Chrome reachable → actionable instruction, non-zero exit, manual path offered.
- Token value is never written to stdout/stderr (assert masked/char-count output only).

### U4. Fix the create date field and verify command/param mapping

**Goal:** Created/edited expenses carry the specified date; command and params are verified correct. (R5, R6)
**Requirements:** R5, R6
**Dependencies:** working auth (U1; U2/U3 help but U1 suffices)
**Files:** `internal/cli/expense_create.go`, `internal/cli/expense_quick.go`, `internal/cli/expense_from_line.go`, `internal/cli/expense_edit.go`, `internal/cli/expense_create_test.go`
**Approach:** Replace the `date` body key with `created` in every create/edit body assembly (the user-facing `--date` flag name stays). Keep amount-in-cents (already correct in all three paths) and `tag`(=department); clarify `--tag` help text to mention it carries the department, and consider a `--department` alias. Verify against the live API that `/RequestMoney` with `created` yields a correctly-dated self-expense with category + department applied; if it does not, route single-create through `Expense_Create` with a one-element `transactionList` (shared with U5) — this is the KTD3 verify-then-pick.
**Execution note:** Start by reproducing the bug with a dry-run diff (body shows `created` not `date`), then confirm a real create lands on the specified date before touching the other call sites.
**Test scenarios:**
- `expense create --merchant X --amount 4900 --date 2025-11-17` → request body contains `created=2025-11-17` (not `date`); dry-run asserts the key.
- Live create (or mock) → expense persists with the given date, not today.
- `quick "Cartesia $49" --date 2025-11-17` → `created` set; amount 4900 cents.
- `from-line` row with an explicit date → `created` carries it.
- `edit --date` → updates via `created`.
- `--tag "101 G&A"` → body `tag=101 G&A`; expense shows that department on read-back.
- Command verification: a created expense round-trips through `sync` and shows the correct `created` and department.

### U5. Bulk-create new expenses via Expense_Create + transactionList

**Goal:** File a list of new expenses in one request. (R7)
**Requirements:** R7
**Dependencies:** U4 (correct per-item field shape, esp. `created`)
**Files:** `internal/cli/expense_bulk.go` (new) or extend `internal/cli/import.go` (its `--batch-size` flag is already stubbed "future: batch API support"), `internal/cli/expense.go` (wire-up), `internal/cli/expense_bulk_test.go` (new)
**Approach:** Assemble an array of transaction objects (`merchant, amount (cents), created, currency, category, tag, reportID, reimbursable, comment`) from a JSONL/CSV input or stdin, post once to `/Expense_Create` with body `{"transactionList": [...]}` (the client JSON-encodes the slice into a single form field — no transport change). Support `--report <reportID>` to attach all items, `--dry-run` to preview the parsed list + total, and a final summary line ("created N expenses, total $X"). Reuse U4's per-item shaping so dates and departments are correct. Decide bulk-vs-existing-`import`: prefer extending `import` so `import expenses --input rows.jsonl --batch` uses `Expense_Create` for one request instead of N.
**Execution note:** Verify a 2-item bulk against the live API before scaling; confirm the reported count matches the report total delta (the manual workaround proved a single bulk request is the reliable shape).
**Test scenarios:**
- JSONL of 3 rows → one POST to `/Expense_Create` whose `transactionList` has 3 correctly-shaped items (dry-run asserts the encoded body).
- Amounts converted to cents; dates under `created`; `tag`=department preserved per row.
- `--report <id>` → every item carries `reportID`.
- `--dry-run` → prints parsed rows + total, sends nothing.
- Empty/whitespace/`#` lines skipped; malformed row reported, others still sent (match `import.go` tolerance).
- Live (or mock) bulk → summary count equals number of items accepted.

### U6. Align doctor, help text, and skill/README docs

**Goal:** A fresh user can self-serve the fixed flows. (R8)
**Requirements:** R8
**Dependencies:** U1–U5
**Files:** `internal/cli/doctor.go`, `internal/cli/auth.go` (help/examples), `cmd/expensify-pp-cli` help text, the published SKILL.md/README for the CLI
**Approach:** `doctor` should distinguish "no token" from "token present but rejected (407)" and point to the right remedy (`auth login`, `auth login --from-chrome`, or pipe a token). Update `auth` examples to show `auth set-token -`, piped stdin, and `--from-chrome`. Correct the SKILL.md auth section (it currently implies `auth set-token` is interactive and omits the stdin/`--from-chrome` paths) and note `tag`=department.
**Test scenarios:**
- `doctor` with empty token → "not authenticated" remedy.
- `doctor` with a present-but-expired token (mock 407) → "session expired, re-auth" remedy, not "ok".
- `--help` for `auth set-token` shows the stdin/`-` usage.
- `Test expectation: docs/help strings — assert examples mention stdin and --from-chrome; no behavioral test beyond doctor classification.`

---

## Risks & Dependencies

- **Financial writes.** `Expense_Create`/`RequestMoney` create real expenses and are not idempotent. Every create/bulk path must keep `--dry-run` and print a verifiable count/total; bulk must never silently retry a partial batch (KTD6).
- **Existing-Chrome capture (U3) is the fragile unit.** App-Bound Encryption (Chrome 127+) blocks naive cookie reads; CDP-attach needs Chrome launched with remote debugging. Mitigation: U1 manual path is the guaranteed fallback and ships first; U3 degrades to instructions, never a hard error.
- **External dependency on `agent-browser`.** The login flow assumes that binary; U2 must keep the no-`agent-browser` fallback intact.
- **`RequestMoney` semantics unverified.** If it ignores `created`/department for self-expenses, U4 pivots single-create onto `Expense_Create` (shares U5). Resolve by live verification, not assumption.
- **Token at rest is plaintext** in `~/.config/expensify-pp-cli/config.toml` (0600); `go-keyring` is a dependency but unused. Keyring-backed storage is out of scope here (see Deferred).

---

## Scope Boundaries

In scope: token acquisition (interactive, piped, existing-Chrome), the `date`→`created` fix and command/param verification across all create/edit paths, and a single-request bulk-create.

### Deferred to Follow-Up Work
- Keyring-backed token storage (replace plaintext config) using the already-present `go-keyring` dep.
- `expense watch` daemon improvements to use the bulk path.
- `report draft` option to bulk-**create** (not just attach local) expenses in range.
- Currency handling beyond USD for bulk rows.

---

## Open Questions

- **PR target / repo state.** The working source at `~/printing-press/library/expensify` is not a git repo, and the expensify CLI was not found under the local `printing-press-library` clones at the documented `library/productivity/expensify` path. Confirm the canonical location/branch and whether to land via printing-press-amend (which opens the PR against `mvanhorn/printing-press-library`) before execution.
- **`/RequestMoney` vs `Expense_Create` for single create.** Decide by live verification in U4 (KTD3).
- **U3 mechanism.** CDP-attach vs cookie-store decryption vs reuse of existing extraction tooling — pick during the U3 spike based on what reliably returns the cookie on macOS.
