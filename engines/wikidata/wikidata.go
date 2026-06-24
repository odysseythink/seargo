package wikidata

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/seargo/seargo/engines/wikimedia"
	"github.com/seargo/seargo/internal/data/externalurls"
	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/internal/httpx"
	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func init() {
	engine.Register("wikidata", &Wikidata{})
}

var sparqlEndpoint = "https://query.wikidata.org/sparql"

type Wikidata struct {
	client *httpx.Client
}

func (w *Wikidata) Name() string { return "wikidata" }

func (w *Wikidata) Categories() []models.Category {
	return []models.Category{models.CategoryGeneral}
}

func (w *Wikidata) Capabilities() engine.Capabilities {
	return engine.Capabilities{SupportsLanguage: true}
}

func (w *Wikidata) About() engine.EngineAbout {
	return engine.EngineAbout{
		Website:    "https://www.wikidata.org",
		WikidataID: "Q2013",
	}
}

func (w *Wikidata) Init(ctx context.Context, cfg engine.EngineInitConfig) bool { return true }

func (w *Wikidata) Setup(cfg engine.EngineInitConfig) bool {
	w.client = cfg.Client
	return true
}

func (w *Wikidata) Search(ctx context.Context, req *models.Request) (*models.Response, error) {
	if w.client == nil {
		return emptyResponse(req), nil
	}

	lang := resolveLanguage(req.Language)
	query, attrs := buildEntitySearchQuery(req.Query, lang)

	httpResp, err := w.client.R().SetContext(ctx).
		SetHeader("Accept", "application/sparql-results+json").
		SetHeader("User-Agent", "SearGo/1.0 (wikidata engine)").
		SetFormData(map[string]string{"query": query}).
		SetTimeout(10 * time.Second).
		Post(sparqlEndpoint)
	if err != nil && httpResp == nil {
		return emptyResponse(req), nil
	}
	if httpResp == nil {
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

		infobox := buildInfoboxFromBinding(b, attrs, lang, entityURL, title)
		return responseWithInfobox(req, &infobox), nil
	}

	return emptyResponse(req), nil
}

func emptyResponse(req *models.Request) *models.Response {
	return &models.Response{
		Query:    req.Query,
		Category: req.Category,
	}
}

func responseWithInfobox(req *models.Request, infobox *results.InfoboxResult) *models.Response {
	typed := []results.Result{infobox}
	raw := make([]any, len(typed))
	for i, r := range typed {
		raw[i] = r
	}
	return &models.Response{
		Query:        req.Query,
		Category:     req.Category,
		Results:      results.ToAPIResult(typed),
		TypedResults: raw,
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

// dummyEntityURLs matches upstream SearXNG.
var dummyEntityURLs = map[string]bool{
	"http://www.wikidata.org/entity/Q4115189":  true,
	"https://www.wikidata.org/entity/Q4115189": true,
	"http://www.wikidata.org/entity/Q13406268": true,
	"https://www.wikidata.org/entity/Q13406268": true,
	"http://www.wikidata.org/entity/Q15397819": true,
	"https://www.wikidata.org/entity/Q15397819": true,
	"http://www.wikidata.org/entity/Q17339402": true,
	"https://www.wikidata.org/entity/Q17339402": true,
}

type wdAttribute struct {
	Name          string
	Label         string
	Type          string // value | label | amount | date | url | image | geo | article
	URLID         string
	URLPathPrefix string
	Priority      int
}

func (a wdAttribute) selectClause() string {
	switch a.Type {
	case "value", "url":
		return "(group_concat(distinct ?" + a.Name + ";separator=', ') as ?" + a.Name + "s)"
	case "label":
		return "(group_concat(distinct ?" + a.Name + "Label;separator=', ') as ?" + a.Name + "Labels)"
	case "image":
		return "(group_concat(distinct ?" + a.Name + ";separator=', ') as ?" + a.Name + "s)"
	case "amount":
		return "?" + a.Name + " ?" + a.Name + "Unit"
	case "date":
		return "?" + a.Name + " ?" + a.Name + "timePrecision ?" + a.Name + "timeZone ?" + a.Name + "timeCalendar"
	case "geo":
		return "?" + a.Name + "Lat ?" + a.Name + "Long"
	case "article":
		return "?article" + a.Name + " ?articleName" + a.Name
	}
	return ""
}

func (a wdAttribute) whereClause(language string) string {
	switch a.Type {
	case "value", "label", "url", "image":
		return "OPTIONAL { ?item wdt:" + a.Name + " ?" + a.Name + ". }"
	case "amount":
		return fmt.Sprintf(`OPTIONAL { ?item p:%[1]s ?%[1]sNode .
  ?%[1]sNode rdf:type wikibase:BestRank ; ps:%[1]s ?%[1]s .
  OPTIONAL { ?%[1]sNode psv:%[1]s/wikibase:quantityUnit ?%[1]sUnit. } }`, a.Name)
	case "date":
		return fmt.Sprintf(`OPTIONAL { ?item p:%[1]s/psv:%[1]s [
    wikibase:timeValue ?%[1]s ;
    wikibase:timePrecision ?%[1]stimePrecision ;
    wikibase:timeTimezone ?%[1]stimeZone ;
    wikibase:timeCalendarModel ?%[1]stimeCalendar ] .
  hint:Prior hint:rangeSafe true. }`, a.Name)
	case "geo":
		return fmt.Sprintf(`OPTIONAL { ?item p:%[1]s/psv:%[1]s [
    wikibase:geoLatitude ?%[1]sLat ;
    wikibase:geoLongitude ?%[1]sLong ] }`, a.Name)
	case "article":
		return fmt.Sprintf(`OPTIONAL { ?article%[1]s schema:about ?item ;
    schema:inLanguage "%[1]s" ;
    schema:isPartOf <https://%[1]s.wikipedia.org/> ;
    schema:name ?articleName%[1]s . }`, a.Name)
	}
	return ""
}

func (a wdAttribute) wikibaseLabel() string {
	if a.Type == "label" {
		return "?" + a.Name + " rdfs:label ?" + a.Name + "Label ."
	}
	return ""
}

func (a wdAttribute) groupBy() string {
	switch a.Type {
	case "amount", "date", "geo", "article":
		return a.selectClause()
	}
	return ""
}

func (a wdAttribute) getStr(b sparqlBinding) string {
	firstValue := func(key string) string {
		v, ok := b[key]
		if !ok {
			return ""
		}
		s := v.Value
		if i := strings.Index(s, ", "); i >= 0 {
			s = s[:i]
		}
		return strings.TrimSpace(s)
	}

	switch a.Type {
	case "value", "url", "image":
		return firstValue(a.Name + "s")
	case "label":
		return firstValue(a.Name + "Labels")
	case "amount":
		val := b[a.Name].Value
		if val == "" {
			return ""
		}
		if unit := b[a.Name+"Unit"].Value; unit != "" {
			val += " " + lastPathSegment(unit)
		}
		return val
	case "date":
		return b[a.Name].Value
	case "geo":
		lat := b[a.Name+"Lat"].Value
		lon := b[a.Name+"Long"].Value
		if lat != "" && lon != "" {
			return lat + " " + lon
		}
	case "article":
		if v, ok := b["article"+a.Name]; ok {
			return v.Value
		}
	}
	return ""
}

func getAttributes(language string) []wdAttribute {
	attrs := []wdAttribute{
		// Dates
		{Name: "P569", Label: "date of birth", Type: "date"},
		{Name: "P570", Label: "date of death", Type: "date"},
		{Name: "P571", Label: "inception", Type: "date"},
		// Labels
		{Name: "P17", Label: "country", Type: "label"},
		{Name: "P27", Label: "country of citizenship", Type: "label"},
		{Name: "P36", Label: "capital", Type: "label"},
		// Values
		{Name: "P1082", Label: "population", Type: "value"},
		// Amounts
		{Name: "P2046", Label: "area", Type: "amount"},
		// URLs
		{Name: "P856", Label: "official website", Type: "url"},
		// Wikipedia article (user language)
		{Name: language, Label: "Wikipedia", Type: "article"},
	}
	if language != "en" {
		attrs = append(attrs, wdAttribute{Name: "en", Label: "Wikipedia", Type: "article"})
	}
	// Geo
	attrs = append(attrs, wdAttribute{Name: "P625", Label: "OpenStreetMap", Type: "geo"})
	// Images - higher Priority wins (matches upstream logic)
	attrs = append(attrs,
		wdAttribute{Name: "P15", Label: "route map", Type: "image", Priority: 1},
		wdAttribute{Name: "P242", Label: "locator map", Type: "image", Priority: 2},
		wdAttribute{Name: "P154", Label: "logo", Type: "image", Priority: 3},
		wdAttribute{Name: "P18", Label: "image", Type: "image", Priority: 4},
		wdAttribute{Name: "P41", Label: "flag", Type: "image", Priority: 5},
		wdAttribute{Name: "P2716", Label: "collage", Type: "image", Priority: 6},
		wdAttribute{Name: "P2910", Label: "icon", Type: "image", Priority: 7},
	)
	return attrs
}

const queryTemplate = `PREFIX wd: <http://www.wikidata.org/entity/>
PREFIX wdt: <http://www.wikidata.org/prop/direct/>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX schema: <http://schema.org/>
PREFIX wikibase: <http://wikiba.se/ontology#>
PREFIX bd: <http://www.bigdata.com/rdf#>
PREFIX mwapi: <https://www.mediawiki.org/ontology#API/>
PREFIX p: <http://www.wikidata.org/prop/>
PREFIX ps: <http://www.wikidata.org/prop/statement/>
PREFIX psv: <http://www.wikidata.org/prop/statement/value/>
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX hint: <http://www.bigdata.com/queryHints#>

SELECT ?item ?itemLabel ?itemDescription %SELECT%
WHERE
{
  SERVICE wikibase:mwapi {
    bd:serviceParam wikibase:endpoint "www.wikidata.org" .
    bd:serviceParam wikibase:api "EntitySearch" .
    bd:serviceParam wikibase:limit 1 .
    bd:serviceParam mwapi:search "%QUERY%" .
    bd:serviceParam mwapi:language "%LANGUAGE%" .
    ?item wikibase:apiOutputItem mwapi:item .
  }
  hint:Prior hint:runFirst "true" .

  %WHERE%

  SERVICE wikibase:label {
    bd:serviceParam wikibase:language "%LANGUAGE%,en" .
    ?item rdfs:label ?itemLabel .
    ?item schema:description ?itemDescription .
    %WIKIBASE_LABELS%
  }
}
GROUP BY ?item ?itemLabel ?itemDescription %GROUP_BY%
`

func buildEntitySearchQuery(query, language string) (string, []wdAttribute) {
	attrs := getAttributes(language)

	var selectParts []string
	var whereParts []string
	var labelParts []string
	var groupByParts []string
	for _, a := range attrs {
		selectParts = append(selectParts, a.selectClause())
		whereParts = append(whereParts, a.whereClause(language))
		if lbl := a.wikibaseLabel(); lbl != "" {
			labelParts = append(labelParts, lbl)
		}
		if gb := a.groupBy(); gb != "" {
			groupByParts = append(groupByParts, gb)
		}
	}

	sparql := queryTemplate
	sparql = strings.ReplaceAll(sparql, "%QUERY%", wikimedia.SparqlEscape(query))
	sparql = strings.ReplaceAll(sparql, "%LANGUAGE%", language)
	sparql = strings.ReplaceAll(sparql, "%SELECT%", strings.Join(selectParts, " "))
	sparql = strings.ReplaceAll(sparql, "%WHERE%", strings.Join(whereParts, "\n  "))
	sparql = strings.ReplaceAll(sparql, "%WIKIBASE_LABELS%", strings.Join(labelParts, "\n      "))
	sparql = strings.ReplaceAll(sparql, "%GROUP_BY%", strings.Join(groupByParts, " "))
	return sparql, attrs
}

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
			Template: "infobox.html",
			Category: string(models.CategoryGeneral),
		},
		InfoboxID: entityURL,
	}

	bestImage := ""
	bestPriority := -1
	articleSet := false

	for _, a := range attrs {
		raw := a.getStr(b)
		if raw == "" {
			continue
		}

		switch a.Type {
		case "image":
			url := externalurls.GetWikimediaThumbnailURL(raw)
			if url != "" && a.Priority > bestPriority {
				bestImage = url
				bestPriority = a.Priority
			}
		case "url":
			url := buildURL(a, raw)
			if url != "" {
				infobox.URLs = append(infobox.URLs, results.InfoboxURL{Title: a.Label, URL: url})
			}
		case "article":
			articleURL := strings.ReplaceAll(raw, "http://", "https://")
			if articleURL != "" {
				infobox.URLs = append(infobox.URLs, results.InfoboxURL{Title: a.Label, URL: articleURL})
				if !articleSet || a.Name != "en" {
					infobox.InfoboxID = articleURL
					articleSet = true
				}
			}
		case "geo":
			lat, lon := parseCoordinate(raw)
			if lat != 0 || lon != 0 {
				zoom := 19
				if areaVal := b["P2046"].Value; areaVal != "" {
					if area, err := strconv.ParseFloat(areaVal, 64); err == nil {
						zoom = externalurls.AreaToOSMZoom(area)
					}
				}
				mapURL := externalurls.GetEarthCoordinatesURL(lat, lon, zoom)
				if mapURL != "" {
					infobox.URLs = append(infobox.URLs, results.InfoboxURL{Title: a.Label, URL: mapURL})
				}
			}
		default:
			value := raw
			if a.Type == "amount" {
				parts := strings.Fields(raw)
				if len(parts) > 0 {
					num := formatByType(parts[0], "amount")
					if len(parts) > 1 {
						num += " " + strings.Join(parts[1:], " ")
					}
					value = num
				}
			} else {
				value = formatByType(raw, a.Type)
			}
			infobox.Attributes = append(infobox.Attributes, results.InfoboxAttribute{
				Label: a.Label,
				Value: value,
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

func buildURL(a wdAttribute, raw string) string {
	raw = strings.TrimSpace(raw)
	if a.URLPathPrefix != "" {
		parts := strings.SplitN(raw, "@", 2)
		if len(parts) == 2 {
			account := strings.TrimSpace(parts[0])
			domain := strings.TrimSpace(parts[1])
			return "https://" + domain + a.URLPathPrefix + account
		}
		return raw
	}
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
	}
	return raw
}

func lastPathSegment(u string) string {
	u = strings.TrimRight(u, "/")
	if i := strings.LastIndex(u, "/"); i >= 0 {
		return u[i+1:]
	}
	return u
}
