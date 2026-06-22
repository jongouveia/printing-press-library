// Copyright 2026 matt-van-horn. Licensed under Apache-2.0.
// Expensify authentication: session login (headed browser), set-token, set-keys,
// status, logout.

package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mvanhorn/printing-press-library/library/productivity/expensify/internal/config"

	"github.com/spf13/cobra"
)

// stdinIsPiped reports whether stdin has piped/redirected data (i.e. is not an
// interactive terminal). Mirrors the isTerminal helper in helpers.go.
func stdinIsPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

// readStdinToken reads all of stdin and returns the trimmed token.
func readStdinToken() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// resolveTokenArg returns a token from a positional arg, or from stdin when the
// arg is "-" or omitted while stdin is piped. Returns "" when no source exists.
func resolveTokenArg(args []string) (string, error) {
	if len(args) == 1 && args[0] != "-" {
		return strings.TrimSpace(args[0]), nil
	}
	if (len(args) == 1 && args[0] == "-") || (len(args) == 0 && stdinIsPiped()) {
		tok, err := readStdinToken()
		if err != nil {
			return "", authErr(fmt.Errorf("reading token from stdin: %w", err))
		}
		return tok, nil
	}
	return "", nil
}

func newAuthCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Expensify session and partner credentials",
	}

	cmd.AddCommand(newAuthStatusCmd(flags))
	cmd.AddCommand(newAuthLoginCmd(flags))
	cmd.AddCommand(newAuthSetTokenCmd(flags))
	cmd.AddCommand(newAuthSetKeysCmd(flags))
	cmd.AddCommand(newAuthLogoutCmd(flags))

	return cmd
}

func newAuthStatusCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show authentication status",
		Example: "  github.com/mvanhorn/printing-press-library/library/productivity/expensify auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			w := cmd.OutOrStdout()
			if !cfg.HasSessionAuth() && !cfg.HasPartnerAuth() {
				fmt.Fprintln(w, red("Not authenticated"))
				fmt.Fprintln(w, "")
				fmt.Fprintln(w, "Log in with a headed browser:")
				fmt.Fprintln(w, "  github.com/mvanhorn/printing-press-library/library/productivity/expensify auth login")
				fmt.Fprintln(w, "Or set manually:")
				fmt.Fprintln(w, "  export EXPENSIFY_AUTH_TOKEN=<token>")
				fmt.Fprintln(w, "  export EXPENSIFY_PARTNER_USER_ID=<id>")
				fmt.Fprintln(w, "  export EXPENSIFY_PARTNER_USER_SECRET=<secret>")
				return authErr(fmt.Errorf("no credentials configured"))
			}
			fmt.Fprintln(w, green("Authenticated"))
			if cfg.HasSessionAuth() {
				fmt.Fprintf(w, "  Session token: present (%d chars)\n", len(cfg.ExpensifyAuthToken))
			}
			if cfg.HasPartnerAuth() {
				fmt.Fprintln(w, "  Partner credentials: set")
			}
			if cfg.AuthSource != "" {
				fmt.Fprintf(w, "  Source: %s\n", cfg.AuthSource)
			}
			fmt.Fprintf(w, "  Config: %s\n", cfg.Path)
			return nil
		},
	}
}

func newAuthLoginCmd(flags *rootFlags) *cobra.Command {
	var sessionName string
	var fallbackToken string
	var loginURL string
	var timeout time.Duration
	var fromChrome bool
	var debugPort int
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in via a headed browser and capture the Expensify session authToken",
		Long: `Launches a headed Chromium browser via 'agent-browser' to www.expensify.com,
waits for you to finish logging in, then reads the authToken cookie and persists it.

Alternatives:
  --from-chrome   read the authToken from your already-signed-in Chrome
  --token <t>     provide the token directly (no browser)
  pipe a token:   pbpaste | github.com/mvanhorn/printing-press-library/library/productivity/expensify auth set-token -

If agent-browser is not installed, falls back to reading the token from stdin.`,
		Example: "  github.com/mvanhorn/printing-press-library/library/productivity/expensify auth login\n  github.com/mvanhorn/printing-press-library/library/productivity/expensify auth login --from-chrome",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			w := cmd.OutOrStdout()

			// Explicit --token overrides and short-circuits the browser flow.
			if fallbackToken != "" {
				if err := cfg.SaveSessionToken(strings.TrimSpace(fallbackToken), ""); err != nil {
					return configErr(err)
				}
				fmt.Fprintf(w, "Token saved to %s\n", cfg.Path)
				return nil
			}

			// --from-chrome: read the token from the user's existing Chrome session.
			if fromChrome {
				token, email, cerr := captureTokenFromChrome(debugPort)
				if cerr != nil {
					return fallbackPromptForToken(cmd, cfg, flags, fmt.Sprintf("could not read from Chrome: %v", cerr))
				}
				if err := cfg.SaveSessionToken(token, email); err != nil {
					return configErr(err)
				}
				fmt.Fprintf(w, "Captured session token (%d chars) from Chrome. Saved to %s\n", len(token), cfg.Path)
				return nil
			}

			if _, err := exec.LookPath("agent-browser"); err != nil {
				return fallbackPromptForToken(cmd, cfg, flags, "agent-browser not found on PATH")
			}

			// Close any stale daemon for this session so --headed can reopen cleanly
			// (re-running login otherwise dead-ends on "daemon already running").
			_ = exec.Command("agent-browser", "--session", sessionName, "close").Run() // #nosec G204 -- fixed binary "agent-browser"; sessionName is a flag passed as a discrete argv element, so there is no shell interpretation or injection surface.

			fmt.Fprintf(w, "Opening %s in session %q ...\n", loginURL, sessionName)
			openCmd := exec.Command("agent-browser", "--session", sessionName, "--headed", "open", loginURL) // #nosec G204 -- fixed binary; sessionName and loginURL are discrete argv elements (no shell), so no injection surface.
			openCmd.Stdout = os.Stderr
			openCmd.Stderr = os.Stderr
			if err := openCmd.Run(); err != nil {
				return fallbackPromptForToken(cmd, cfg, flags, fmt.Sprintf("failed to open browser: %v", err))
			}

			fmt.Fprintln(w, "Log in to Expensify in the opened browser window.")
			fmt.Fprintf(w, "Polling for the authToken cookie every 5 seconds (timeout %s).\n", timeout)

			deadline := time.Now().Add(timeout)
			for time.Now().Before(deadline) {
				token, email, err := captureTokenViaAgentBrowser(sessionName)
				if err == nil && token != "" {
					if err := cfg.SaveSessionToken(token, email); err != nil {
						return configErr(err)
					}
					if email != "" {
						fmt.Fprintf(w, "Captured session token (%d chars) for %s. Saved to %s\n", len(token), email, cfg.Path)
					} else {
						fmt.Fprintf(w, "Captured session token (%d chars). Saved to %s\n", len(token), cfg.Path)
					}
					return nil
				}
				time.Sleep(5 * time.Second)
			}
			return fallbackPromptForToken(cmd, cfg, flags, fmt.Sprintf("login poll timed out after %s", timeout))
		},
	}
	cmd.Flags().StringVar(&sessionName, "session", "expensify-pp-login", "Named agent-browser session to use")
	cmd.Flags().StringVar(&fallbackToken, "token", "", "Provide the authToken directly instead of opening a browser")
	cmd.Flags().StringVar(&loginURL, "site", "https://www.expensify.com/", "Login URL to open (classic www by default; new.expensify.com also works)")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "How long to poll for the authToken cookie")
	cmd.Flags().BoolVar(&fromChrome, "from-chrome", false, "Read the authToken from your already-signed-in Chrome instead of opening a browser")
	cmd.Flags().IntVar(&debugPort, "chrome-debug-port", 0, "If set, read cookies from a Chrome started with --remote-debugging-port=<port>")
	return cmd
}

// captureTokenViaAgentBrowser runs `agent-browser cookies get --json` and
// extracts the authToken cookie value (and associated account email if any).
func captureTokenViaAgentBrowser(session string) (string, string, error) {
	out, err := exec.Command("agent-browser", "--session", session, "cookies", "get", "--json").Output() // #nosec G204 -- fixed binary; session is a discrete argv element (no shell), so no injection surface.
	if err != nil {
		return "", "", err
	}
	// Output is expected to be a JSON array. Be defensive: accept {cookies: [...]}
	// as well.
	var asArray []map[string]any
	if err := json.Unmarshal(out, &asArray); err != nil {
		var wrapper struct {
			Cookies []map[string]any `json:"cookies"`
		}
		if err := json.Unmarshal(out, &wrapper); err != nil {
			return "", "", fmt.Errorf("parsing cookies JSON: %w", err)
		}
		asArray = wrapper.Cookies
	}
	var token, email string
	for _, c := range asArray {
		name, _ := c["name"].(string)
		domain, _ := c["domain"].(string)
		if !strings.Contains(strings.ToLower(domain), "expensify.com") {
			continue
		}
		val, _ := c["value"].(string)
		switch name {
		case "authToken":
			token = val
		case "email", "userEmail", "expensify_email":
			email = val
		}
	}
	if token == "" {
		return "", "", fmt.Errorf("authToken cookie not present yet")
	}
	return token, email, nil
}

func fallbackPromptForToken(cmd *cobra.Command, cfg *config.Config, flags *rootFlags, reason string) error {
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Could not auto-capture session: %s\n", reason)

	manualHint := "Provide a token directly with: auth login --token <t>, " +
		"or pipe it: pbpaste | auth set-token -"

	// If stdin is piped, accept the token from it — even under --no-input.
	if stdinIsPiped() {
		tok, rerr := readStdinToken()
		if rerr == nil && tok != "" {
			if err := cfg.SaveSessionToken(tok, ""); err != nil {
				return configErr(err)
			}
			fmt.Fprintf(w, "Token read from stdin. Saved to %s\n", cfg.Path)
			return nil
		}
		return authErr(fmt.Errorf("no token on stdin. %s", manualHint))
	}

	if flags.noInput {
		return authErr(fmt.Errorf("cannot prompt for token with --no-input. %s", manualHint))
	}

	fmt.Fprintln(w, "Paste your authToken (find it at https://www.expensify.com/ -> DevTools -> Application -> Cookies -> authToken):")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	token := strings.TrimSpace(line)
	if token == "" {
		return authErr(fmt.Errorf("no token provided. %s", manualHint))
	}
	if err := cfg.SaveSessionToken(token, ""); err != nil {
		return configErr(err)
	}
	fmt.Fprintf(w, "Token saved to %s\n", cfg.Path)
	return nil
}

func newAuthSetTokenCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "set-token [token|-]",
		Short: "Save a session authToken to the config file",
		Long: `Save a session authToken. Provide it as an argument, as "-" to read it
from stdin, or pipe it on stdin with no argument:

  github.com/mvanhorn/printing-press-library/library/productivity/expensify auth set-token <800-char-hex>
  pbpaste | github.com/mvanhorn/printing-press-library/library/productivity/expensify auth set-token -
  pbpaste | github.com/mvanhorn/printing-press-library/library/productivity/expensify auth set-token`,
		Example: "  pbpaste | github.com/mvanhorn/printing-press-library/library/productivity/expensify auth set-token -",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			token, err := resolveTokenArg(args)
			if err != nil {
				return err
			}
			if token == "" {
				return authErr(fmt.Errorf("no token provided (pass it as an argument, \"-\", or pipe it on stdin)"))
			}
			if err := cfg.SaveSessionToken(token, ""); err != nil {
				return configErr(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Token saved to %s\n", cfg.Path)
			return nil
		},
	}
}

func newAuthSetKeysCmd(flags *rootFlags) *cobra.Command {
	var fromEnv bool
	var partnerID, partnerSecret string
	cmd := &cobra.Command{
		Use:     "set-keys",
		Short:   "Save partnerUserID and partnerUserSecret for Integration Server calls",
		Example: "  github.com/mvanhorn/printing-press-library/library/productivity/expensify auth set-keys --env\n  github.com/mvanhorn/printing-press-library/library/productivity/expensify auth set-keys --id <id> --secret <secret>",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			if fromEnv {
				partnerID = os.Getenv("EXPENSIFY_PARTNER_USER_ID")
				partnerSecret = os.Getenv("EXPENSIFY_PARTNER_USER_SECRET")
			}
			if partnerID == "" || partnerSecret == "" {
				if flags.noInput {
					return authErr(fmt.Errorf("partner id/secret missing and --no-input set; pass --id/--secret or set env vars"))
				}
				reader := bufio.NewReader(os.Stdin)
				if partnerID == "" {
					fmt.Fprint(cmd.OutOrStdout(), "partnerUserID: ")
					s, _ := reader.ReadString('\n')
					partnerID = strings.TrimSpace(s)
				}
				if partnerSecret == "" {
					fmt.Fprint(cmd.OutOrStdout(), "partnerUserSecret: ")
					s, _ := reader.ReadString('\n')
					partnerSecret = strings.TrimSpace(s)
				}
			}
			if partnerID == "" || partnerSecret == "" {
				return authErr(fmt.Errorf("partnerUserID and partnerUserSecret are both required"))
			}
			cfg.ExpensifyPartnerUserId = partnerID
			cfg.ExpensifyPartnerUserSecret = partnerSecret
			if err := cfg.SaveTokens(cfg.ClientID, cfg.ClientSecret, cfg.AccessToken, cfg.RefreshToken, cfg.TokenExpiry); err != nil {
				return configErr(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Partner credentials saved to %s\n", cfg.Path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&fromEnv, "env", false, "Read EXPENSIFY_PARTNER_USER_ID and EXPENSIFY_PARTNER_USER_SECRET from the environment")
	cmd.Flags().StringVar(&partnerID, "id", "", "partnerUserID")
	cmd.Flags().StringVar(&partnerSecret, "secret", "", "partnerUserSecret")
	return cmd
}

func newAuthLogoutCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:     "logout",
		Short:   "Clear stored credentials",
		Example: "  github.com/mvanhorn/printing-press-library/library/productivity/expensify auth logout",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			cfg.ExpensifyAuthToken = ""
			cfg.ExpensifyPartnerUserId = ""
			cfg.ExpensifyPartnerUserSecret = ""
			if err := cfg.ClearTokens(); err != nil {
				return configErr(err)
			}
			if os.Getenv("EXPENSIFY_AUTH_TOKEN") != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Config cleared. Note: EXPENSIFY_AUTH_TOKEN env var is still set.")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out. Credentials cleared.")
			return nil
		},
	}
}
