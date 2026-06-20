package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
