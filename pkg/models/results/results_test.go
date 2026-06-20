package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseResult_Defaults(t *testing.T) {
	br := BaseResult{
		Title:    "Test Title",
		URL:      "https://example.com",
		Content:  "Test snippet",
		Engine:   "google",
		Template: "default",
	}

	assert.Equal(t, "Test Title", br.Title)
	assert.Equal(t, "https://example.com", br.URL)
	assert.Equal(t, "default", br.Template)
}

func TestMainResult(t *testing.T) {
	mr := &MainResult{
		BaseResult: BaseResult{
			Title:    "Main Result",
			URL:      "https://example.com/page",
			Content:  "Description",
			Engine:   "bing",
			Template: "default",
		},
	}

	assert.Equal(t, "Main Result", mr.Title)
	assert.Equal(t, "default", mr.Template)
}

func TestImageResult_Fields(t *testing.T) {
	ir := &ImageResult{
		BaseResult: BaseResult{
			Title:     "An image",
			URL:       "https://example.com/img",
			Template:  "images.html",
		},
		ThumbnailSrc: "https://example.com/thumb.jpg",
		ImgSrc:       "https://example.com/full.jpg",
		Resolution:   "1920x1080",
	}

	assert.Equal(t, "images.html", ir.Template)
	assert.Equal(t, "1920x1080", ir.Resolution)
	assert.False(t, ir.IsBase64()) // not base64 data
}

func TestImageResult_IsBase64(t *testing.T) {
	ir := &ImageResult{
		ImgSrc: "data:image/png;base64,iVBORw0KGgo=",
	}
	assert.True(t, ir.IsBase64())
}

func TestVideoResult_Fields(t *testing.T) {
	vr := &VideoResult{
		BaseResult: BaseResult{
			Title:    "Video",
			Template: "videos.html",
		},
		Thumbnail: "https://example.com/thumb.jpg",
		IFrameSrc: "https://example.com/embed",
		Length:    "3:45",
	}

	assert.Equal(t, "videos.html", vr.Template)
	assert.Equal(t, "3:45", vr.Length)
}

func TestResultType_InterfaceSatisfaction(t *testing.T) {
	var r Result = &MainResult{}
	assert.NotNil(t, r)

	r = &ImageResult{}
	assert.NotNil(t, r)

	r = &VideoResult{}
	assert.NotNil(t, r)

	r = &NewsResult{}
	assert.NotNil(t, r)

	r = &PaperResult{}
	assert.NotNil(t, r)
}
