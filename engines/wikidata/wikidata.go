package wikidata

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/seargo/seargo/internal/data/externalurls"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func init() {
	engine.Register("wikidata", &Wikidata{})
}

// sparqlEndpoint is the Wikidata SPARQL endpoint. It is a variable so tests can
// override it with a local mock server.
var sparqlEndpoint = "https://query.wikidata.org/sparql"

// queryTemplate is the SPARQL query used to search for entities and fetch a
// fixed catalog of properties. Placeholders are %QUERY%, %LANGUAGE%, %SELECT%,
// and %WHERE%.
const queryTemplate = `PREFIX wd: <http://www.wikidata.org/entity/>
PREFIX wdt: <http://www.wikidata.org/prop/direct/>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX schema: <http://schema.org/>
PREFIX wikibase: <http://wikiba.se/ontology#>
PREFIX bd: <http://www.bigdata.com/rdf#>
PREFIX mwapi: <https://www.mediawiki.org/ontology#API/>

SELECT ?item ?itemLabel ?itemDescription%SELECT%
WHERE {
  SERVICE wikibase:mwapi {
    bd:serviceParam wikibase:api "EntitySearch" .
    bd:serviceParam wikibase:endpoint "www.wikidata.org" .
    bd:serviceParam mwapi:search "%QUERY%" .
    bd:serviceParam mwapi:language "%LANGUAGE%" .
    ?item wikibase:apiOutputItem mwapi:item .
  }
  SERVICE wikibase:label { bd:serviceParam wikibase:language "%LANGUAGE%,en". }
%WHERE%
}
LIMIT 10
`

// dummyEntityURLs are Wikidata entities that should never be returned as results.
var dummyEntityURLs = map[string]bool{
	"http://www.wikidata.org/entity/Q16958":  true, // Wikimedia disambiguation page
	"https://www.wikidata.org/entity/Q16958": true,
}

// Wikidata queries Wikidata SPARQL and returns InfoboxResult values.
type Wikidata struct {
	client *httpx.Client
}

// Name returns the engine name.
func (w *Wikidata) Name() string { return "wikidata" }

// Categories returns the categories this engine serves.
func (w *Wikidata) Categories() []models.Category {
	return []models.Category{models.CategoryGeneral}
}

// Capabilities describes engine features.
func (w *Wikidata) Capabilities() engine.Capabilities {
	return engine.Capabilities{
		SupportsLanguage: true,
	}
}

// Init performs any asynchronous initialization.
func (w *Wikidata) Init(ctx context.Context, cfg engine.EngineInitConfig) bool {
	return true
}

// Setup receives the engine's runtime configuration.
func (w *Wikidata) Setup(cfg engine.EngineInitConfig) bool {
	w.client = cfg.Client
	return true
}

// About returns descriptive metadata for the engine.
func (w *Wikidata) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://www.wikidata.org",
		WikidataID: "Q2013",
	}
}

// Search queries Wikidata SPARQL and returns infobox results.
func (w *Wikidata) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	if w.client == nil {
		return emptyResponse(req), nil
	}

	language := resolveLanguage(req.Language)
	query, attrs := buildEntitySearchQuery(req.Query, language)

	httpResp, err := w.client.R().SetContext(ctx).
		SetHeader("Accept", "application/sparql-results+json").
		SetHeader("User-Agent", "SearGo/1.0 (wikidata engine)").
		SetFormData(map[string]string{"query": query}).
		SetTimeout(10 * time.Second).
		Post(sparqlEndpoint)
	if err != nil {
		return emptyResponse(req), nil
	}
	if httpResp.StatusCode != 200 {
		return emptyResponse(req), nil
	}

	var sparqlResp sparqlResponse
	if err := json.Unmarshal(httpResp.Body, &sparqlResp); err != nil {
		return emptyResponse(req), nil
	}

	seen := make(map[string]bool)
	var typed []results.Result
	for _, b := range sparqlResp.Results.Bindings {
		item := b["item"]
		entityURL := item.Value
		if entityURL == "" || seen[entityURL] || dummyEntityURLs[entityURL] {
			continue
		}
		seen[entityURL] = true

		title := b["itemLabel"].Value
		if title == "" {
			continue
		}

		infobox := buildInfoboxFromBinding(b, attrs, language, entityURL, title)
		// Prefer the Wikipedia article URL as the canonical infobox ID so that
		// Wikidata and Wikipedia results for the same entity can be merged.
		if articleURL := extractRaw(b, "article"); articleURL != "" {
			infobox.InfoboxID = articleURL
		}
		typed = append(typed, &infobox)
	}

	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
		Results:  results.ToAPIResult(typed),
	}, nil
}

func emptyResponse(req *models.Request) *models.Response {
	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
		Results:  nil,
	}
}

func resolveLanguage(lang string) string {
	if lang == "" {
		return "en"
	}
	lang = strings.SplitN(lang, "-", 2)[0]
	lang = strings.SplitN(lang, "_", 2)[0]
	if lang == "" {
		return "en"
	}
	return lang
}

// wdAttribute describes one Wikidata property to fetch and how to present it.
type wdAttribute struct {
	PID      string // Wikidata property id, e.g. "P18", or "_wikipedia" for article link
	Label    string // human label, e.g. "image"
	Type     string // "label" | "value" | "amount" | "date" | "url" | "image" | "geo"
	URLID    string // key in data/external_urls.json (for url/image)
	Priority int    // for images: lower is more preferred
}

func (a wdAttribute) valueVar() string {
	if a.PID == "_wikipedia" {
		return "article"
	}
	if a.Type == "label" {
		return a.PID + "l"
	}
	return a.PID + "v"
}

func (a wdAttribute) selectClause() string {
	if a.PID == "_wikipedia" {
		return " ?article"
	}
	if a.Type == "label" {
		return " ?" + a.PID + "v ?" + a.PID + "l"
	}
	return " ?" + a.PID + "v"
}

func (a wdAttribute) whereClause(language string) string {
	if a.PID == "_wikipedia" {
		return fmt.Sprintf(`  OPTIONAL {
    ?article schema:about ?item ;
             schema:inLanguage "%s" ;
             schema:isPartOf <https://%s.wikipedia.org/> .
  }`, language, language)
	}
	if a.Type == "label" {
		return fmt.Sprintf(`  OPTIONAL {
    ?item wdt:%s ?%sv.
    OPTIONAL { ?%sv rdfs:label ?%sl. FILTER(LANG(?%sl) = "%s") }
  }`, a.PID, a.PID, a.PID, a.PID, a.PID, language)
	}
	return fmt.Sprintf(`  OPTIONAL { ?item wdt:%s ?%sv. }`, a.PID, a.PID)
}

// getAttributes returns the fixed catalog of Wikidata properties to fetch.
func getAttributes() []wdAttribute {
	return []wdAttribute{
		// dates
		{PID: "P571", Label: "inception", Type: "date"},
		{PID: "P569", Label: "date of birth", Type: "date"},
		{PID: "P570", Label: "date of death", Type: "date"},
		// labels
		{PID: "P27", Label: "country of citizenship", Type: "label"},
		{PID: "P17", Label: "country", Type: "label"},
		{PID: "P36", Label: "capital", Type: "label"},
		{PID: "P37", Label: "official language", Type: "label"},
		// values
		{PID: "P1082", Label: "population", Type: "value"},
		{PID: "P498", Label: "currency code", Type: "value"},
		// amounts
		{PID: "P2046", Label: "area", Type: "amount"},
		{PID: "P2048", Label: "height", Type: "amount"},
		// URLs
		{PID: "P856", Label: "official website", Type: "url", URLID: ""},
		{PID: "P345", Label: "IMDb", Type: "url", URLID: "imdb_id"},
		{PID: "P2002", Label: "Twitter", Type: "url", URLID: "twitter_profile"},
		{PID: "P2013", Label: "Facebook", Type: "url", URLID: "facebook_profile"},
		// images (priority lower = more preferred)
		{PID: "P18", Label: "image", Type: "image", URLID: "wikimedia_image", Priority: 4},
		{PID: "P154", Label: "logo", Type: "image", URLID: "wikimedia_image", Priority: 3},
		{PID: "P41", Label: "flag", Type: "image", URLID: "wikimedia_image", Priority: 5},
		// geo
		{PID: "P625", Label: "coordinate location", Type: "geo"},
		// wikipedia article
		{PID: "_wikipedia", Label: "Wikipedia", Type: "url"},
	}
}

func buildEntitySearchQuery(query, language string) (string, []wdAttribute) {
	attrs := getAttributes()

	var selectParts []string
	var whereParts []string
	for _, a := range attrs {
		selectParts = append(selectParts, a.selectClause())
		whereParts = append(whereParts, a.whereClause(language))
	}

	sparql := queryTemplate
	sparql = strings.ReplaceAll(sparql, "%QUERY%", sparqlEscape(query))
	sparql = strings.ReplaceAll(sparql, "%LANGUAGE%", language)
	sparql = strings.ReplaceAll(sparql, "%SELECT%", strings.Join(selectParts, ""))
	sparql = strings.ReplaceAll(sparql, "%WHERE%", strings.Join(whereParts, "\n"))
	return sparql, attrs
}

func sparqlEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

// sparqlResponse mirrors the JSON returned by the Wikidata SPARQL endpoint.
type sparqlResponse struct {
	Results struct {
		Bindings []sparqlBinding `json:"bindings"`
	} `json:"results"`
}

type sparqlBinding map[string]struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	DataType string `json:"datatype,omitempty"`
}

func buildInfoboxFromBinding(b sparqlBinding, attrs []wdAttribute, language, entityURL, title string) results.InfoboxResult {
	infobox := results.InfoboxResult{
		BaseResult: results.BaseResult{
			Title:    title,
			Content:  b["itemDescription"].Value,
			Engine:   "wikidata",
			Template: "infobox",
			Category: string(models.CategoryGeneral),
		},
		InfoboxID: entityURL,
	}

	bestImage := ""
	bestPriority := math.MaxInt32

	for _, a := range attrs {
		raw := extractValue(b, a)
		if raw == "" {
			continue
		}

		switch a.Type {
		case "image":
			url := wikimediaThumbnailURL(raw)
			if url != "" && a.Priority < bestPriority {
				bestImage = url
				bestPriority = a.Priority
			}
		case "url":
			url := buildURL(a, raw)
			if url != "" {
				infobox.URLs = append(infobox.URLs, results.InfoboxURL{Title: a.Label, URL: url})
			}
		case "geo":
			lat, lon := parseCoordinate(raw)
			if lat != 0 || lon != 0 {
				zoom := 19
				if areaVal := extractRaw(b, "P2046v"); areaVal != "" {
					if area, err := strconv.ParseFloat(areaVal, 64); err == nil {
						zoom = externalurls.AreaToOSMZoom(area)
					}
				}
				mapURL := externalurls.GetEarthCoordinatesURL(lat, lon, zoom)
				if mapURL != "" {
					infobox.URLs = append(infobox.URLs, results.InfoboxURL{Title: "OpenStreetMap", URL: mapURL})
				}
			}
		default:
			infobox.Attributes = append(infobox.Attributes, results.InfoboxAttribute{
				Label: a.Label,
				Value: formatByType(raw, a.Type),
			})
		}
	}

	if bestImage != "" {
		infobox.ImgSrc = bestImage
		infobox.ImgAlt = title
	}

	infobox.URLs = append(infobox.URLs, results.InfoboxURL{Title: "Wikidata", URL: entityURL})

	return infobox
}

func extractValue(b sparqlBinding, a wdAttribute) string {
	key := a.valueVar()
	if v, ok := b[key]; ok && v.Value != "" {
		return v.Value
	}
	// For label-type properties, fall back to the Q-id from the value variable.
	if a.Type == "label" {
		if v, ok := b[a.PID+"v"]; ok && v.Value != "" {
			return lastPathSegment(v.Value)
		}
	}
	return ""
}

func extractRaw(b sparqlBinding, key string) string {
	if v, ok := b[key]; ok {
		return v.Value
	}
	return ""
}

func lastPathSegment(u string) string {
	u = strings.TrimRight(u, "/")
	if i := strings.LastIndex(u, "/"); i >= 0 {
		return u[i+1:]
	}
	return u
}

func formatByType(raw, typ string) string {
	switch typ {
	case "date":
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			return t.Format("2006-01-02")
		}
		if idx := strings.Index(raw, "T"); idx >= 0 {
			return raw[:idx]
		}
		return raw
	case "amount":
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			return strconv.FormatFloat(f, 'f', -1, 64)
		}
		return raw
	default:
		return raw
	}
}

func wikimediaThumbnailURL(raw string) string {
	filename := externalurls.GetWikimediaImageID(raw)
	if filename == "" {
		return ""
	}
	return "https://commons.wikimedia.org/wiki/Special:FilePath/" + url.PathEscape(filename) + "?width=300"
}

func buildURL(a wdAttribute, raw string) string {
	if a.URLID == "" {
		return raw
	}
	if a.URLID == "imdb_id" {
		urlID := externalurls.GetIMDBURLID(raw)
		if urlID != "" {
			return externalurls.GetExternalURL(urlID, raw, "")
		}
		return raw
	}
	return externalurls.GetExternalURL(a.URLID, raw, "")
}

func parseCoordinate(raw string) (lat, lon float64) {
	raw = strings.TrimPrefix(raw, "Point(")
	raw = strings.TrimSuffix(raw, ")")
	parts := strings.Fields(raw)
	if len(parts) != 2 {
		return 0, 0
	}
	lon, _ = strconv.ParseFloat(parts[0], 64)
	lat, _ = strconv.ParseFloat(parts[1], 64)
	return lat, lon
}
