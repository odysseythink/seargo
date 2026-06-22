package render

import (
	"encoding/json"

	"github.com/seargo/seargo/pkg/models"
)

// JSONWriter renders search results as JSON.
type JSONWriter struct{}

// Render marshals the response to JSON bytes.
func (w *JSONWriter) Render(resp *models.Response) ([]byte, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ContentType returns the MIME type for JSON.
func (w *JSONWriter) ContentType() string {
	return "application/json; charset=utf-8"
}
