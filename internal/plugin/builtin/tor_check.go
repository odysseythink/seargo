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
		return &torCheckPlugin{
			httpClient: &http.Client{Timeout: 5 * time.Second},
		}
	})
}

type torCheckPlugin struct {
	httpClient *http.Client
}

func (t *torCheckPlugin) ID() string { return "tor_check" }

func (t *torCheckPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                t.ID(),
		Name:              "Tor Check",
		Description:       "Check if your IP address is a known Tor exit node.",
		PreferenceSection: "privacy",
		Keywords:          []string{"tor-check", "tor_check", "torcheck", "tor", "tor check"},
	}
}

func (t *torCheckPlugin) Init(ctx *plugin.AppContext) bool { return true }
func (t *torCheckPlugin) PreSearch(ctx *plugin.SearchContext) bool { return true }
func (t *torCheckPlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool { return true }

func (t *torCheckPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	if ctx.PageNo > 1 {
		return nil
	}

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

	remoteAddr, _ := ctx.Preferences["remote_addr"].(string)
	if remoteAddr == "" {
		remoteAddr = "unknown"
	}

	isTor, err := t.checkTorExitNode(remoteAddr)
	if err != nil {
		return []models.Result{
			{Kind: "answer", Title: "Tor check unavailable", Content: "Tor check unavailable"},
		}
	}

	var answer string
	if isTor {
		answer = fmt.Sprintf("Your IP %s appears to be a Tor exit node.", remoteAddr)
	} else {
		answer = fmt.Sprintf("Your IP %s does not appear to be a Tor exit node.", remoteAddr)
	}
	return []models.Result{
		{Kind: "answer", Title: answer, Content: answer},
	}
}

func (t *torCheckPlugin) checkTorExitNode(ip string) (bool, error) {
	if t.httpClient == nil {
		return false, fmt.Errorf("no HTTP client configured")
	}
	resp, err := t.httpClient.Get("https://check.torproject.org/exit-addresses")
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
