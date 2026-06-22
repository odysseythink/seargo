package render

import (
	"encoding/xml"
	"net/url"
	"time"

	"github.com/seargo/seargo/pkg/models"
)

// RSSWriter renders search results as an Atom 1.0 feed.
type RSSWriter struct{}

type atomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string      `xml:"title"`
	Link    atomLink    `xml:"link"`
	Updated string      `xml:"updated"`
	ID      string      `xml:"id"`
	Entry   []atomEntry `xml:"entry"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
}

type atomEntry struct {
	Title   string     `xml:"title"`
	Link    atomLink   `xml:"link"`
	Summary string     `xml:"summary"`
	Updated string     `xml:"updated"`
	ID      string     `xml:"id"`
	Author  atomAuthor `xml:"author"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

const maxQueryLen = 64

// Render writes the response as an Atom feed XML.
func (w *RSSWriter) Render(resp *models.Response, baseURL string) ([]byte, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	truncated := truncateQuery(resp.Query, maxQueryLen)
	feed := atomFeed{
		Title:   "Search: " + truncated,
		Link:    atomLink{Href: searchURL(baseURL, resp.Query), Rel: "self"},
		Updated: now,
		ID:      searchURL(baseURL, resp.Query),
	}
	for _, r := range resp.Results {
		updated := now
		if r.PublishedAt != nil {
			updated = r.PublishedAt.Format(time.RFC3339)
		}
		entry := atomEntry{
			Title:   r.Title,
			Link:    atomLink{Href: r.URL},
			Summary: truncateContent(r.Content, 500),
			Updated: updated,
			ID:      r.URL,
			Author:  atomAuthor{Name: r.Engine},
		}
		feed.Entry = append(feed.Entry, entry)
	}
	data, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), data...), nil
}

// ContentType returns the MIME type for Atom.
func (w *RSSWriter) ContentType() string {
	return "application/atom+xml; charset=utf-8"
}

func searchURL(baseURL, query string) string {
	return baseURL + "/search?q=" + url.QueryEscape(query)
}

func truncateQuery(q string, maxLen int) string {
	runes := []rune(q)
	if len(runes) <= maxLen {
		return q
	}
	return string(runes[:maxLen]) + "..."
}

func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen]) + "..."
}
