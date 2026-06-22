# Expensify Printed CLI Agent Guide

This directory is a generated `expensify-pp-cli` printed CLI. It was produced by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press), so treat systemic fixes as upstream Printing Press fixes first. Keep local edits narrow and document why a generated-tree patch belongs here.

## Local Operating Contract

Start by asking the generated CLI for current runtime truth:

```bash
expensify-pp-cli doctor --json
expensify-pp-cli agent-context --pretty
```

Use runtime discovery instead of relying on a copied command list:

```bash
expensify-pp-cli <command> --help
```

Add `--agent` to command invocations for JSON, compact output, non-interactive defaults, no color, and confirmation-safe scripting:

```bash
expensify-pp-cli <command> --agent
```

Before running an unfamiliar command that may mutate remote state, inspect its help and prefer a dry run:

```bash
expensify-pp-cli <command> --help
expensify-pp-cli <command> --dry-run --agent
```

Use `--yes --no-input` only after the target, arguments, and side effects are clear.

## Auth

Expensify session tokens are short-lived. Acquire a token with `auth login --from-chrome` (reads the freshest `www.expensify.com` authToken from your signed-in Chrome) or pipe one in: `pbpaste | expensify-pp-cli auth set-token -`. The CLI also honors the `EXPENSIFY_AUTH_TOKEN` env var. Run `doctor` to distinguish "no token" from "token present but expired (407)".

For install, examples, and longer product guidance, read `README.md` and `SKILL.md`. This file intentionally stays small so repo-local agents get invariant local guidance without duplicating the generated docs.

## Local Customizations

This CLI carries hand-edits beyond generator output; each is recorded in `.printing-press-patches.json` at this CLI's root so the change isn't lost on the next regen and is visible to the next reader. That file is an index of customizations, not a second copy of the diff — diffs live in `git`.
