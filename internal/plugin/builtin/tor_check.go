package builtin

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	plugin.RegisterBuiltin("tor_check", func() plugin.Plugin {
		return &torCheckPlugin{}
	})
}

// torCheckPlugin adds a result showing whether the user's IP is a known Tor exit node.
// It runs via PostSearch when the query matches one of its keywords exactly.
type torCheckPlugin struct {
	httpClient *http.Client // optional; defaults to 5s timeout client
}

// getHTTPClient returns the configured client or a default one.
func (t *torCheckPlugin) getHTTPClient() *http.Client {
	if t.httpClient != nil {
		return t.httpClient
	}
	return &http.Client{Timeout: 5 * time.Second}
}

func (t *torCheckPlugin) ID() string { return "tor_check" }

func (t *torCheckPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                t.ID(),
		Name:              "Tor Check",
		Description:       "Check if your IP address is a known Tor exit node. Triggered by 'tor-check' keyword.",
		PreferenceSection: "general",
		Keywords:          []string{"tor-check", "tor_check", "torcheck", "tor", "tor check"},
	}
}

func (t *torCheckPlugin) Init(ctx *plugin.AppContext) bool {
	return true
}

func (t *torCheckPlugin) PreSearch(ctx *plugin.SearchContext) bool {
	return true
}

func (t *torCheckPlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool {
	return true
}

func (t *torCheckPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	if ctx.PageNo > 1 {
		return nil
	}

	// Check exact keyword match: the query trimmed and lowered must match one of the keywords exactly.
	query := strings.TrimSpace(strings.ToLower(ctx.Query))
	matched := false
	for _, kw := range t.Info().Keywords {
		if query == strings.ToLower(kw) {
			matched = true
			break
		}
	}
	if !matched {
		return nil
	}

	ip := ctx.RemoteAddr
	if ip == "" {
		ip = "unknown"
	}

	// Fetch Tor exit node list and check the IP.
	isTor, err := checkTorExitNode(t.getHTTPClient(), ip)
	if err != nil {
		return []models.Result{
			{
				Kind:    "answer",
				Title:   "Tor Check",
				Content: "Tor check unavailable",
				Engine:  "tor_check",
			},
		}
	}

	result := models.Result{
		Kind:   "answer",
		Title:  "Tor Check",
		Engine: "tor_check",
	}

	if isTor {
		result.Content = fmt.Sprintf("Your IP %s appears to be a Tor exit node.", ip)
	} else {
		result.Content = fmt.Sprintf("Your IP %s does not appear to be a Tor exit node.", ip)
	}

	return []models.Result{result}
}

// checkTorExitNode fetches the Tor exit list and checks if the given IP is listed.
func checkTorExitNode(client *http.Client, ip string) (bool, error) {
	resp, err := client.Get("https://check.torproject.org/exit-addresses")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ExitAddress ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 && parts[1] == ip {
				return true, nil
			}
		}
	}

	return false, scanner.Err()
}
