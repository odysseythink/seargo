package bases

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seargo/seargo/pkg/models"
	"github.com/seargo/seargo/pkg/models/results"
)

func TestBuildTypedResult_PaperFromJSON(t *testing.T) {
	raw := map[string]interface{}{
		"doi":       "10.1000/182",
		"journal":   "Nature",
		"authors":   []interface{}{"Alice", "Bob"},
		"pdf_url":   "https://example.com/paper.pdf",
		"published": "2024-01-15",
	}
	cfg := ResultTypeConfig{
		Type:               ResultTypePaper,
		DOIQuery:           "doi",
		JournalQuery:       "journal",
		AuthorsQuery:       "authors",
		PDFURLQuery:        "pdf_url",
		PublishedDateQuery: "published",
	}
	base := models.Result{Title: "T", URL: "https://example.com", Content: "C", Engine: "x", Category: models.CategoryScience}

	r := buildTypedResult(raw, cfg, base)
	pr, ok := r.(*results.PaperResult)
	require.True(t, ok)
	assert.Equal(t, "paper", pr.Kind())
	assert.Equal(t, "10.1000/182", pr.DOI)
	assert.Equal(t, "Nature", pr.Journal)
	assert.Equal(t, []string{"Alice", "Bob"}, pr.Authors)
	assert.Equal(t, "https://example.com/paper.pdf", pr.PDFURL)
}

func TestBuildTypedResult_FileFromJSON(t *testing.T) {
	raw := map[string]interface{}{
		"mime":   "application/pdf",
		"size":   "2048",
		"name":   "doc.pdf",
		"magnet": "magnet:?xt=urn:btih:abc",
	}
	cfg := ResultTypeConfig{
		Type:           ResultTypeFile,
		FileTypeQuery:  "mime",
		FileSizeQuery:  "size",
		FilenameQuery:  "name",
		MagnetURIQuery: "magnet",
	}
	base := models.Result{Title: "T", URL: "https://example.com", Engine: "x", Category: models.CategoryFiles}

	r := buildTypedResult(raw, cfg, base)
	fr, ok := r.(*results.FileResult)
	require.True(t, ok)
	assert.Equal(t, "file", fr.Kind())
	assert.Equal(t, "application/pdf", fr.FileType)
	assert.Equal(t, int64(2048), fr.FileSize)
	assert.Equal(t, "doc.pdf", fr.Filename)
	assert.Equal(t, "magnet:?xt=urn:btih:abc", fr.MagnetURI)
}

func TestJSONQuery_DollarText(t *testing.T) {
	data := map[string]interface{}{
		"title": map[string]interface{}{
			"$":    "Real Title",
			"type": "literal",
		},
	}
	assert.Equal(t, "Real Title", firstString(jsonQuery(data, "title/$")))
}
