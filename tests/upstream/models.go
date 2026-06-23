package upstream

import "encoding/json"

// UpstreamResult mirrors the SearXNG JSON result item.
type UpstreamResult struct {
	URL           string          `json:"url"`
	Title         string          `json:"title"`
	Content       string          `json:"content"`
	Engine        string          `json:"engine"`
	Engines       []string        `json:"engines"`
	Category      string          `json:"category"`
	Template      string          `json:"template"`
	Score         float64         `json:"score"`
	Positions     []int           `json:"positions"`
	Thumbnail     string          `json:"thumbnail"`
	ImgSrc        string          `json:"img_src"`
	PublishedDate string          `json:"publishedDate"`
	Extra         json.RawMessage `json:"-"`
}

// UpstreamAnswer mirrors an upstream answer item.
type UpstreamAnswer struct {
	Answer  string `json:"answer"`
	URL     string `json:"url,omitempty"`
	Content string `json:"content,omitempty"`
}

// UpstreamInfobox mirrors an upstream infobox item.
type UpstreamInfobox struct {
	Title      string          `json:"title"`
	URL        string          `json:"url,omitempty"`
	Content    string          `json:"content,omitempty"`
	Engine     string          `json:"engine,omitempty"`
	Engines    []string        `json:"engines,omitempty"`
	ImgSrc     string          `json:"img_src,omitempty"`
	Attributes json.RawMessage `json:"attributes,omitempty"`
	URLs       json.RawMessage `json:"urls,omitempty"`
}

// UpstreamUnresponsiveEngine mirrors upstream failure metadata.
type UpstreamUnresponsiveEngine struct {
	Engine    string `json:"engine"`
	Reason    string `json:"reason"`
	Suspended bool   `json:"suspended,omitempty"`
}

// UnmarshalJSON handles both object {engine, reason} and array [engine, reason] forms.
func (u *UpstreamUnresponsiveEngine) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == '[' {
		var arr []string
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		if len(arr) > 0 {
			u.Engine = arr[0]
		}
		if len(arr) > 1 {
			u.Reason = arr[1]
		}
		return nil
	}
	type alias UpstreamUnresponsiveEngine
	return json.Unmarshal(data, (*alias)(u))
}

// UpstreamResponse mirrors the SearXNG /search?format=json response.
type UpstreamResponse struct {
	Query               string                       `json:"query"`
	Results             []UpstreamResult             `json:"results"`
	Answers             []UpstreamAnswer             `json:"answers"`
	Corrections         []string                     `json:"corrections"`
	Infoboxes           []UpstreamInfobox            `json:"infoboxes"`
	Suggestions         []string                     `json:"suggestions"`
	UnresponsiveEngines []UpstreamUnresponsiveEngine `json:"unresponsive_engines"`
}
