package builtin

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/seargo/seargo/internal/plugin"
	"github.com/seargo/seargo/pkg/models"
)

type hashPlugin struct{}

var hashPattern = regexp.MustCompile(`^(?i)(md5|sha1|sha224|sha256|sha384|sha512)\s+(.+)`)

func init() {
	plugin.RegisterBuiltin("hash_plugin", func() plugin.Plugin { return &hashPlugin{} })
}

func (p *hashPlugin) ID() string { return "hash_plugin" }

func (p *hashPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		ID:                "hash_plugin",
		Name:              "Hash Plugin",
		Description:       "Compute hash values of strings using various algorithms (MD5, SHA-1, SHA-2)",
		PreferenceSection: "query",
		Keywords:          []string{"md5", "sha1", "sha224", "sha256", "sha384", "sha512"},
		Examples:          []string{"md5 hello world", "sha256 hello"},
	}
}

func (p *hashPlugin) Init(ctx *plugin.AppContext) bool                { return true }
func (p *hashPlugin) PreSearch(ctx *plugin.SearchContext) bool       { return true }
func (p *hashPlugin) OnResult(ctx *plugin.SearchContext, r *models.Result) bool { return true }

func (p *hashPlugin) PostSearch(ctx *plugin.SearchContext) []models.Result {
	matches := hashPattern.FindStringSubmatch(ctx.Query)
	if len(matches) < 3 {
		return nil
	}

	algo := strings.ToLower(matches[1])
	input := matches[2]

	var hash string
	switch algo {
	case "md5":
		h := md5.Sum([]byte(input))
		hash = hex.EncodeToString(h[:])
	case "sha1":
		h := sha1.Sum([]byte(input))
		hash = hex.EncodeToString(h[:])
	case "sha224":
		h := sha256.Sum224([]byte(input))
		hash = hex.EncodeToString(h[:])
	case "sha256":
		h := sha256.Sum256([]byte(input))
		hash = hex.EncodeToString(h[:])
	case "sha384":
		h := sha512.Sum384([]byte(input))
		hash = hex.EncodeToString(h[:])
	case "sha512":
		h := sha512.Sum512([]byte(input))
		hash = hex.EncodeToString(h[:])
	default:
		return nil
	}

	return []models.Result{{
		Kind:    "answer",
		Title:   fmt.Sprintf("%s(%q) = %s", algo, input, hash),
		Content: hash,
		Engine:  "hash_plugin",
	}}
}
