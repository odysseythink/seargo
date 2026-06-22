package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/seargo/seargo/cmd/seargo-update/internal"
)

const defaultReleasesURL = "https://ftp.mozilla.org/pub/firefox/releases/"

var versionRE = regexp.MustCompile(`^(\d+)\.(\d)(?:\.\d)?$`)

type version struct {
	Major int
	Minor int
}

type userAgentData struct {
	OS       []string `json:"os"`
	UA       string   `json:"ua"`
	Versions []string `json:"versions"`
}

func main() {
	var (
		out         = flag.String("out", "data/useragents.json", "output JSON path")
		releasesURL = flag.String("releases-url", defaultReleasesURL, "Mozilla Firefox releases URL")
	)
	flag.Parse()

	if err := Run(*out, nil, *releasesURL); err != nil {
		fmt.Fprintf(os.Stderr, "update-useragents: %v\n", err)
		os.Exit(1)
	}
}

// Run fetches Firefox release versions and writes data/useragents.json.
func Run(outPath string, client fetch.Client, releasesURL string) error {
	h := fetch.New(client)
	ctx := context.Background()

	body, err := h.Get(ctx, releasesURL)
	if err != nil {
		return fmt.Errorf("fetch releases: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("parse releases HTML: %w", err)
	}

	versions := []version{}
	releasePath := "/pub/firefox/releases/"
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		if !strings.HasPrefix(href, releasePath) {
			return
		}
		verStr := strings.Trim(href[len(releasePath):], "/")
		match := versionRE.FindStringSubmatch(verStr)
		if match == nil {
			return
		}
		major, _ := strconv.Atoi(match[1])
		minor, _ := strconv.Atoi(match[2])
		versions = append(versions, version{Major: major, Minor: minor})
	})

	if len(versions) == 0 {
		return fmt.Errorf("no firefox versions found")
	}

	sort.Slice(versions, func(i, j int) bool {
		if versions[i].Major != versions[j].Major {
			return versions[i].Major > versions[j].Major
		}
		return versions[i].Minor > versions[j].Minor
	})

	latestMajor := versions[0].Major
	keep := map[int]bool{latestMajor: true, latestMajor - 1: true}

	result := []string{}
	seen := map[string]bool{}
	for _, v := range versions {
		if !keep[v.Major] {
			continue
		}
		sv := fmt.Sprintf("%d.%d", v.Major, v.Minor)
		if seen[sv] {
			continue
		}
		seen[sv] = true
		result = append(result, sv)
	}

	ua := userAgentData{
		Versions: result,
		OS: []string{
			"Windows NT 10.0; Win64; x64",
			"X11; Linux x86_64",
			"Macintosh; Intel Mac OS X 10.15",
			"Macintosh; Intel Mac OS X 11.0",
		},
		UA: "Mozilla/5.0 ({os}; rv:{version}) Gecko/20100101 Firefox/{version}",
	}

	return writeJSON(outPath, ua)
}

func writeJSON(outPath string, ua userAgentData) error {
	enc, err := json.MarshalIndent(ua, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	enc = append(enc, '\n')

	tmp := outPath + ".tmp"
	if err := os.WriteFile(tmp, enc, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	return os.Rename(tmp, outPath)
}
