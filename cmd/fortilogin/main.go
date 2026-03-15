package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

const (
	baseURL              = "https://firewall.nitkkr.ac.in:1000"
	connectivityURL      = "http://connectivitycheck.gstatic.com/generate_204"
	loginProbeURL        = "http://neverssl.com"
	fixedLogoutToken     = "080a080d06020649"
	daemonInterval       = 30 * time.Second
	alertInterval        = 30 * time.Second
	maxCredentialFailure = 2
)

type Config struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	FailedAttempts int    `json:"failed_attempts"`
	PortalBaseURL  string `json:"portal_base_url,omitempty"`
}

type NetworkState string

const (
	StateOnline   NetworkState = "online"
	StateFirewall NetworkState = "firewall_blocked"
	StateOffline  NetworkState = "offline"
)

var (
	appConfig Config
	httpAgent *http.Client
)

func main() {
	initHTTPClient()

	command := "daemon"
	if len(os.Args) > 1 {
		command = strings.ToLower(os.Args[1])
	}

	if err := loadConfig(); err != nil {
		fatalf("failed to load config: %v", err)
	}

	switch command {
	case "daemon":
		ensureCredentials()
		runDaemon()
	case "login":
		ensureCredentials()
		exitForResult(runLogin())
	case "logout":
		exitForResult(runLogout())
	case "update":
		promptAndSaveCredentials()
		logf("credentials updated")
	case "status":
		exitForResult(runStatus())
	case "help", "--help", "-h":
		printHelp()
	default:
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("fortilogin <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  daemon   Keep checking connectivity and auto-login when firewall blocks internet")
	fmt.Println("  login    Trigger one login attempt immediately")
	fmt.Println("  logout   Logout the active firewall session without printing HTML")
	fmt.Println("  update   Update saved credentials and reset failure state")
	fmt.Println("  status   Show current network and credential status")
}

func getConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	path := filepath.Join(configDir, "fortilogin")
	_ = os.MkdirAll(path, 0o700)
	return filepath.Join(path, "config.json")
}

func getLegacyConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "NitAgent", "config.json")
}

func loadConfig() error {
	data, err := os.ReadFile(getConfigPath())
	if errors.Is(err, os.ErrNotExist) {
		legacyData, legacyErr := os.ReadFile(getLegacyConfigPath())
		if errors.Is(legacyErr, os.ErrNotExist) {
			appConfig = Config{}
			return nil
		}
		if legacyErr != nil {
			return legacyErr
		}
		if err := json.Unmarshal(legacyData, &appConfig); err != nil {
			return err
		}
		return saveConfig()
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &appConfig)
}

func saveConfig() error {
	data, err := json.MarshalIndent(appConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigPath(), data, 0o600)
}

func ensureCredentials() {
	if strings.TrimSpace(appConfig.Username) != "" && strings.TrimSpace(appConfig.Password) != "" {
		return
	}
	logf("no saved credentials found")
	promptAndSaveCredentials()
}

func promptAndSaveCredentials() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Roll Number (Username): ")
	appConfig.Username = mustReadLine(reader)

	fmt.Print("Enter Password: ")
	appConfig.Password = mustReadLine(reader)
	appConfig.FailedAttempts = 0

	if err := saveConfig(); err != nil {
		fatalf("failed to save credentials: %v", err)
	}
}

func mustReadLine(reader *bufio.Reader) string {
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		fatalf("failed reading input: %v", err)
	}
	return strings.TrimSpace(line)
}

func initHTTPClient() {
	jar, _ := cookiejar.New(nil)
	httpAgent = &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:    32,
			MaxConnsPerHost: 32,
		},
		Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func runDaemon() {
	logf("daemon started")

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var previous NetworkState = "unknown"
	ticker := time.NewTicker(daemonInterval)
	defer ticker.Stop()

	for {
		state := detectNetworkState()
		if state != previous {
			logf("network state: %s", state)
			previous = state
		}

		switch state {
		case StateOnline:
		case StateOffline:
		case StateFirewall:
			if appConfig.FailedAttempts >= maxCredentialFailure {
				alertInvalidCredentials(sigCtx)
				return
			}
			result := attemptLogin()
			handleLoginResult(result)
			if result == "wrong_credentials" && appConfig.FailedAttempts >= maxCredentialFailure {
				alertInvalidCredentials(sigCtx)
				return
			}
		}

		select {
		case <-sigCtx.Done():
			logf("daemon stopped")
			return
		case <-ticker.C:
		}
	}
}

func runLogin() int {
	if detectNetworkState() == StateOnline {
		logf("already online; no login needed")
		return 0
	}

	if appConfig.FailedAttempts >= maxCredentialFailure {
		logf("credentials are locked after %d failures; run `fortilogin update`", maxCredentialFailure)
		return 1
	}

	result := attemptLogin()
	handleLoginResult(result)
	if result == "success" {
		return 0
	}
	return 1
}

func runLogout() int {
	ok, err := logoutSession(fixedLogoutToken)
	if err != nil {
		logf("logout failed: %v", err)
		return 1
	}
	if ok {
		logf("successful logout")
		return 0
	}

	logf("logout response did not confirm success")
	return 1
}

func runStatus() int {
	credState := "configured"
	if strings.TrimSpace(appConfig.Username) == "" || strings.TrimSpace(appConfig.Password) == "" {
		credState = "missing"
	}
	if appConfig.FailedAttempts >= maxCredentialFailure {
		credState = "locked"
	}

	fmt.Printf("credentials: %s\n", credState)
	fmt.Printf("failed_attempts: %d\n", appConfig.FailedAttempts)
	fmt.Printf("network: %s\n", detectNetworkState())
	fmt.Println("logout: fixed-token-supported")
	return 0
}

func detectNetworkState() NetworkState {
	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(connectivityURL)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNoContent {
			return StateOnline
		}
	}

	if _, err := net.LookupHost("firewall.nitkkr.ac.in"); err != nil {
		return StateOffline
	}

	req, _ := http.NewRequest(http.MethodGet, loginProbeURL, nil)
	resp, err = client.Do(req)
	if err != nil {
		return StateOffline
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		if strings.Contains(location, "firewall.nitkkr.ac.in") {
			return StateFirewall
		}
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if strings.Contains(string(body), "firewall.nitkkr.ac.in") || strings.Contains(string(body), "magic") {
		return StateFirewall
	}

	return StateOffline
}

func attemptLogin() string {
	headers := map[string]string{
		"User-Agent":                "Mozilla/5.0",
		"Upgrade-Insecure-Requests": "1",
	}

	req, _ := http.NewRequest(http.MethodGet, loginProbeURL, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := httpAgent.Do(req)
	if err != nil {
		return "network_error"
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	redirectURL := extractRedirectURL(body)
	magicToken := extractMagicToken(body)
	dynamicBaseURL := strings.TrimSpace(appConfig.PortalBaseURL)

	if redirectURL != "" {
		if strings.HasPrefix(redirectURL, "http://") {
			redirectURL = strings.Replace(redirectURL, "http://", "https://", 1)
		}
		if parsedBase := extractBaseURL(redirectURL); parsedBase != "" {
			dynamicBaseURL = parsedBase
			appConfig.PortalBaseURL = parsedBase
			_ = saveConfig()
		}
		if queryIndex := strings.Index(redirectURL, "?"); queryIndex != -1 {
			magicToken = redirectURL[queryIndex+1:]
		}

		req, _ = http.NewRequest(http.MethodGet, redirectURL, nil)
		req.Header.Set("Referer", redirectURL)
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err = httpAgent.Do(req)
		if err == nil {
			defer resp.Body.Close()
			bodyBytes, _ = io.ReadAll(resp.Body)
			if magicToken == "" {
				magicToken = extractMagicToken(string(bodyBytes))
			}
		}
	}

	if magicToken == "" {
		return "no_token"
	}

	if dynamicBaseURL == "" {
		dynamicBaseURL = baseURL
	}

	form := url.Values{}
	form.Set("username", appConfig.Username)
	form.Set("password", appConfig.Password)
	form.Set("magic", magicToken)
	form.Set("4Tredir", "https://google.com")

	loginURL := fmt.Sprintf("%s/auth", dynamicBaseURL)
	req, _ = http.NewRequest(http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err = httpAgent.Do(req)
	if err != nil {
		return "network_error"
	}
	defer resp.Body.Close()

	authBody, _ := io.ReadAll(resp.Body)
	authText := string(authBody)

	switch {
	case strings.Contains(authText, "keepalive"):
		appConfig.FailedAttempts = 0
		_ = saveConfig()
		return "success"
	case strings.Contains(authText, "Authentication Failed"), resp.StatusCode == http.StatusUnauthorized:
		return "wrong_credentials"
	case strings.Contains(authText, "Maximum number of"), strings.Contains(authText, "simultaneous"):
		return "limit_reached"
	default:
		return "unknown_failure"
	}
}

func extractRedirectURL(body string) string {
	re := regexp.MustCompile(`window\.location="([^"]+)"`)
	match := re.FindStringSubmatch(body)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func extractMagicToken(body string) string {
	re := regexp.MustCompile(`name="magic" value="([a-fA-F0-9]+)"`)
	match := re.FindStringSubmatch(body)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func extractBaseURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
}

func handleLoginResult(result string) {
	switch result {
	case "success":
		appConfig.FailedAttempts = 0
		_ = saveConfig()
		logf("logged in successfully")
	case "wrong_credentials":
		appConfig.FailedAttempts++
		_ = saveConfig()
		logf("authentication failed (%d/%d); run `fortilogin update`", appConfig.FailedAttempts, maxCredentialFailure)
	case "limit_reached":
		logf("session limit reached; logout another device or run `fortilogin logout` here")
	case "network_error":
		logf("network error while trying to login")
	case "no_token":
		logf("firewall login page did not expose a usable token")
	default:
		logf("login failed: %s", result)
	}
}

func alertInvalidCredentials(ctx context.Context) {
	ticker := time.NewTicker(alertInterval)
	defer ticker.Stop()

	logf("credentials failed %d times; daemon is paused until `fortilogin update` is run", maxCredentialFailure)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logf("credentials failed %d times; run `fortilogin update` as soon as possible", maxCredentialFailure)
		}
	}
}

func logoutSession(token string) (bool, error) {
	var lastErr error
	for _, candidate := range candidatePortalBaseURLs() {
		reqURL := fmt.Sprintf("%s/logout?%s", candidate, token)
		resp, err := httpAgent.Get(reqURL)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		text := strings.ToLower(string(body))
		return strings.Contains(text, "successfully logged out"), nil
	}
	if lastErr != nil {
		return false, lastErr
	}
	return false, errors.New("no reachable firewall portal URL is known")
}

func candidatePortalBaseURLs() []string {
	seen := map[string]bool{}
	var bases []string

	for _, candidate := range []string{strings.TrimSpace(appConfig.PortalBaseURL), baseURL} {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		bases = append(bases, candidate)
	}

	return bases
}

func exitForResult(code int) {
	os.Exit(code)
}

func logf(format string, args ...any) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[fatal] %s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}
