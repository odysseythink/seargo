package upstream

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpstreamResponse_Unmarshal(t *testing.T) {
	payload := `{
		"query": "golang",
		"results": [
			{
				"url": "https://go.dev/doc/",
				"title": "Documentation - The Go Programming Language",
				"content": "The Go programming language ...",
				"engine": "google",
				"engines": ["google"],
				"score": 1.0,
				"category": "general",
				"template": "default.html",
				"thumbnail": "",
				"positions": [1]
			}
		],
		"answers": [],
		"corrections": [],
		"infoboxes": [],
		"suggestions": ["golang tutorial"],
		"unresponsive_engines": []
	}`

	var resp UpstreamResponse
	err := json.Unmarshal([]byte(payload), &resp)
	require.NoError(t, err)
	require.Equal(t, "golang", resp.Query)
	require.Len(t, resp.Results, 1)
	require.Equal(t, "https://go.dev/doc/", resp.Results[0].URL)
	require.Equal(t, "default.html", resp.Results[0].Template)
	require.Equal(t, []string{"golang tutorial"}, resp.Suggestions)
}
