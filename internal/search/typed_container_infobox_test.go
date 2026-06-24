package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func TestNormalizeInfoboxID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://www.wikidata.org/entity/Q64", "Q64"},
		{"http://www.wikidata.org/entity/Q64", "Q64"},
		{"https://en.wikipedia.org/wiki/Berlin", "Berlin"},
		{"https://de.wikipedia.org/wiki/Berlin", "Berlin"},
		{"https://en.wikipedia.org/wiki/Berlin/", "Berlin"},
		{"https://en.wikipedia.org/wiki/Berlin_(city)", "Berlin_(city)"},
		{"https://example.com/entity/Q64", "https://example.com/entity/Q64"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeInfoboxID(tt.input))
		})
	}
}

func TestTypedContainer_MergeInfoboxByEntityID(t *testing.T) {
	c := NewTypedResultContainer(nil)

	wikidataResult := models.Result{
		Kind:     "infobox",
		Title:    "Berlin",
		Content:  "Capital and largest city of Germany",
		URL:      "https://www.wikidata.org/entity/Q64",
		Engine:   "wikidata",
		Category: models.CategoryGeneral,
		Extra: map[string]any{
			"infobox_id": "https://en.wikipedia.org/wiki/Berlin",
			"img_src":    "https://commons.wikimedia.org/wiki/Special:FilePath/Berlin.jpg?width=300",
			"attributes": []results.InfoboxAttribute{
				{Label: "population", Value: "3669495"},
			},
			"urls": []results.InfoboxURL{
				{Title: "Wikidata", URL: "https://www.wikidata.org/entity/Q64"},
			},
		},
	}

	wikipediaResult := models.Result{
		Kind:     "infobox",
		Title:    "Berlin",
		Content:  "Capital of Germany",
		URL:      "https://en.wikipedia.org/wiki/Berlin",
		Engine:   "wikipedia",
		Category: models.CategoryGeneral,
		Extra: map[string]any{
			"infobox_id": "https://en.wikipedia.org/wiki/Berlin",
			"attributes": []results.InfoboxAttribute{
				{Label: "Country", Value: "Germany"},
			},
			"urls": []results.InfoboxURL{
				{Title: "Wikipedia", URL: "https://en.wikipedia.org/wiki/Berlin"},
			},
		},
	}

	c.Extend("wikidata", []models.Result{wikidataResult}, 0)
	c.Extend("wikipedia", []models.Result{wikipediaResult}, 0)

	infos := c.GetInfoboxes()
	require.Len(t, infos, 1, "infoboxes with the same normalized entity id should merge")

	info := infos[0]
	assert.Equal(t, "Berlin", info.Title)
	assert.ElementsMatch(t, []string{"wikidata", "wikipedia"}, info.Engines)
	assert.Equal(t, "https://commons.wikimedia.org/wiki/Special:FilePath/Berlin.jpg?width=300", info.ImgSrc)

	require.Len(t, info.Attributes, 2)
	labels := []string{info.Attributes[0].Label, info.Attributes[1].Label}
	assert.ElementsMatch(t, []string{"population", "Country"}, labels)

	require.Len(t, info.URLs, 2)
	urlTitles := []string{info.URLs[0].Title, info.URLs[1].Title}
	assert.ElementsMatch(t, []string{"Wikidata", "Wikipedia"}, urlTitles)
}

func TestTypedContainer_AddInfoboxesMerge(t *testing.T) {
	c := NewTypedResultContainer(nil)

	c.AddInfoboxes("wikidata", []models.Infobox{
		{
			InfoboxID: "https://en.wikipedia.org/wiki/Berlin",
			Title:     "Berlin",
			URL:       "https://www.wikidata.org/entity/Q64",
			Content:   "Capital and largest city of Germany",
			Engine:    "wikidata",
			ImgSrc:    "https://commons.wikimedia.org/wiki/Special:FilePath/Berlin.jpg?width=300",
			Attributes: []models.InfoboxAttribute{
				{Label: "population", Value: "3669495"},
			},
			URLs: []models.InfoboxURL{
				{Title: "Wikidata", URL: "https://www.wikidata.org/entity/Q64"},
			},
		},
	})

	c.AddInfoboxes("wikipedia", []models.Infobox{
		{
			Title:   "Berlin",
			URL:     "https://en.wikipedia.org/wiki/Berlin",
			Content: "Capital of Germany",
			Engine:  "wikipedia",
			Attributes: []models.InfoboxAttribute{
				{Label: "Country", Value: "Germany"},
			},
			URLs: []models.InfoboxURL{
				{Title: "Wikipedia", URL: "https://en.wikipedia.org/wiki/Berlin"},
			},
		},
	})

	infos := c.GetInfoboxes()
	require.Len(t, infos, 1)
	assert.ElementsMatch(t, []string{"wikidata", "wikipedia"}, infos[0].Engines)
	assert.Len(t, infos[0].Attributes, 2)
	assert.Len(t, infos[0].URLs, 2)
}

func TestTypedContainer_InfoboxNotInResults(t *testing.T) {
	c := NewTypedResultContainer(map[string]float64{})
	c.Extend("wikidata", []models.Result{
		{
			Kind:    "infobox",
			Title:   "Berlin",
			URL:     "https://en.wikipedia.org/wiki/Berlin",
			Content: "Capital of Germany",
			Engine:  "wikidata",
			Extra:   map[string]any{"infobox_id": "https://en.wikipedia.org/wiki/Berlin"},
		},
		{
			Kind:     "main",
			Title:    "Berlin tourism",
			URL:      "https://example.com/berlin",
			Engine:   "some-engine",
			Category: models.CategoryGeneral,
		},
	}, 0)

	results := c.Results()
	require.Len(t, results, 1)
	assert.Equal(t, "main", results[0].Kind)

	infoboxes := c.GetInfoboxes()
	require.Len(t, infoboxes, 1)
	assert.Equal(t, "Berlin", infoboxes[0].Title)
}
