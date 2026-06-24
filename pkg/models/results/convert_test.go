package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToAPIResult_MainResult(t *testing.T) {
	mr := &MainResult{
		BaseResult: BaseResult{
			Title:    "Search Result",
			URL:      "https://example.com/page",
			Content:  "A description",
			Engine:   "google",
			Template: "default",
		},
	}

	apiResults := ToAPIResult([]Result{mr})
	assert.Len(t, apiResults, 1)

	r := apiResults[0]
	assert.Equal(t, "Search Result", r.Title)
	assert.Equal(t, "https://example.com/page", r.URL)
	assert.Equal(t, "A description", r.Content)
	assert.Equal(t, "google", r.Engine)
	assert.Equal(t, "main", r.Kind)
	assert.Equal(t, "default.html", r.Template)
}

func TestToAPIResult_ImageResult(t *testing.T) {
	ir := &ImageResult{
		BaseResult: BaseResult{
			Title:    "Cat Photo",
			URL:      "https://example.com/cat",
			Engine:   "google",
			Template: "images.html",
		},
		ImgSrc:       "https://example.com/cat.jpg",
		ThumbnailSrc: "https://example.com/cat_thumb.jpg",
		Resolution:   "800x600",
	}

	apiResults := ToAPIResult([]Result{ir})
	assert.Len(t, apiResults, 1)

	r := apiResults[0]
	assert.Equal(t, "image", r.Kind)
	assert.Equal(t, "images.html", r.Template)
	assert.Equal(t, "https://example.com/cat_thumb.jpg", r.ThumbnailURL)
	assert.Equal(t, "https://example.com/cat.jpg", r.Extra["img_src"])
}

func TestToAPIResult_VideoResult(t *testing.T) {
	vr := &VideoResult{
		BaseResult: BaseResult{
			Title:    "Video",
			URL:      "https://example.com/video",
			Template: "videos.html",
		},
		Thumbnail: "https://example.com/thumb.jpg",
		Length:    "5:00",
	}

	apiResults := ToAPIResult([]Result{vr})
	assert.Len(t, apiResults, 1)
	assert.Equal(t, "video", apiResults[0].Kind)
	assert.Equal(t, "videos.html", apiResults[0].Template)
}

func TestToAPIResult_MultipleTypes(t *testing.T) {
	results := []Result{
		&MainResult{BaseResult: BaseResult{Title: "Main", URL: "https://a.com", Template: "default", Engine: "g"}},
		&ImageResult{BaseResult: BaseResult{Title: "Img", URL: "https://b.com", Template: "images.html", Engine: "g"}},
		&NewsResult{BaseResult: BaseResult{Title: "News", URL: "https://c.com", Template: "default", Engine: "g"}},
	}

	apiResults := ToAPIResult(results)
	assert.Len(t, apiResults, 3)
}

func TestToAPIResult_EmptyInput(t *testing.T) {
	apiResults := ToAPIResult(nil)
	assert.Nil(t, apiResults)

	apiResults = ToAPIResult([]Result{})
	assert.Nil(t, apiResults)
}

func TestToAPIResult_FileResult(t *testing.T) {
	fr := &FileResult{
		BaseResult: BaseResult{
			Title:  "sample.pdf",
			URL:    "https://example.com/sample.pdf",
			Engine: "wikicommons",
		},
	}
	api := ToAPIResult([]Result{fr})
	assert.Len(t, api, 1)
	assert.Equal(t, "file", api[0].Kind)
	assert.Equal(t, "files.html", api[0].Template)
}

func TestToAPIResult_EngineDataPassthrough(t *testing.T) {
	mr := &MainResult{
		BaseResult: BaseResult{
			Title:      "T",
			URL:        "https://x.com",
			EngineData: map[string]any{"key": "value"},
		},
	}

	apiResults := ToAPIResult([]Result{mr})
	assert.NotNil(t, apiResults[0].Extra)
	assert.Equal(t, "value", apiResults[0].Extra["key"])
}

func TestToAPIResult_CodeResult(t *testing.T) {
	cr := &CodeResult{
		BaseResult: BaseResult{
			Title:  "main.go",
			URL:    "https://github.com/foo/bar/blob/main.go",
			Engine: "github_code",
		},
		Repository:   "foo/bar",
		CodeLanguage: "go",
	}
	api := ToAPIResult([]Result{cr})
	require.Len(t, api, 1)
	assert.Equal(t, "code", api[0].Kind)
	assert.Equal(t, "code.html", api[0].Template)
	assert.Equal(t, "go", api[0].Extra["code_language"])
}

func TestToAPIResult_PaperResult(t *testing.T) {
	pr := &PaperResult{
		BaseResult: BaseResult{
			Title:  "A Sample Paper",
			URL:    "https://example.com/paper",
			Engine: "openairepublications",
		},
		DOI: "10.1000/xyz",
	}
	api := ToAPIResult([]Result{pr})
	require.Len(t, api, 1)
	assert.Equal(t, "paper", api[0].Kind)
	assert.Equal(t, "paper.html", api[0].Template)
	assert.Equal(t, "10.1000/xyz", api[0].Extra["doi"])
}

func TestToAPIResult_KeyValueResult(t *testing.T) {
	kv := &KeyValueResult{
		BaseResult: BaseResult{
			Title:  "Albert Einstein",
			Engine: "wikidata",
		},
		KVMap: map[string]string{
			"date of birth": "1879-03-14",
		},
	}
	api := ToAPIResult([]Result{kv})
	require.Len(t, api, 1)
	assert.Equal(t, "keyvalue", api[0].Kind)
	assert.Equal(t, "keyvalue.html", api[0].Template)
	assert.Equal(t, "1879-03-14", api[0].Extra["kv_map"].(map[string]string)["date of birth"])
}
