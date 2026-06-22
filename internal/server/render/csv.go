package render

import (
	"bytes"
	"encoding/csv"
	"strconv"
	"strings"
	"time"

	"github.com/seargo/seargo/pkg/models"
)

// CSVWriter renders search results as CSV.
// Cells starting with =, +, -, @ are prefixed with a single quote to prevent
// spreadsheet formula injection.
type CSVWriter struct{}

var csvHeader = []string{"title", "url", "content", "engine", "score", "published_at"}

// Render writes the response as CSV bytes.
func (w *CSVWriter) Render(resp *models.Response) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write(csvHeader); err != nil {
		return nil, err
	}
	for _, r := range resp.Results {
		row := []string{
			sanitizeCSV(r.Title),
			sanitizeCSV(r.URL),
			sanitizeCSV(r.Content),
			r.Engine,
			formatScore(r.Score),
			formatTime(r.PublishedAt),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ContentType returns the MIME type for CSV.
func (w *CSVWriter) ContentType() string {
	return "text/csv; charset=utf-8"
}

func sanitizeCSV(s string) string {
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "=") || strings.HasPrefix(s, "+") ||
		strings.HasPrefix(s, "-") || strings.HasPrefix(s, "@") {
		return "'" + s
	}
	return s
}

func formatScore(score float64) string {
	return strconv.FormatFloat(score, 'f', -1, 64)
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
