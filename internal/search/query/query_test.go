package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/pkg/models"
)

func TestTimeoutParser(t *testing.T) {
	p := TimeoutParser{}
	tests := []struct {
		input    string
		wantOK   bool
		wantSecs float64
	}{
		{"<3", true, 3.0},
		{"<850", true, 0.85},
		{"<100", true, 0.1},
		{"<0", true, 0.0},
		{"golang", false, 0},
		{"<", false, 0},
		{"<abc", false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ok := p.Check(tt.input)
			assert.Equal(t, tt.wantOK, ok)
			if ok {
				part, parsed := p.Parse(tt.input)
				assert.True(t, parsed)
				var secs float64
				fmt.Sscanf(part.Value, "%f", &secs)
				assert.Equal(t, tt.wantSecs, secs)
			}
		})
	}
}

func TestLanguageParser_DirectCode(t *testing.T) {
	p := LanguageParser{}
	tests := []struct {
		input  string
		wantOK bool
		want   string
	}{
		{":en", true, "en"},
		{":zh-CN", true, "zh-CN"},
		{":zh-cn", true, "zh-CN"},
		{":en-US", true, "en-US"},
		{":en_us", true, "en-US"},
		{":EN", true, "en"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.wantOK, p.Check(tt.input))
			if tt.wantOK {
				part, ok := p.Parse(tt.input)
				assert.True(t, ok)
				assert.Equal(t, tt.want, part.Value)
			}
		})
	}
}

func TestLanguageParser_NameMapping(t *testing.T) {
	p := LanguageParser{}
	tests := []struct {
		input  string
		wantOK bool
		want   string
	}{
		{":english", true, "en"},
		{":german", true, "de"},
		{":french", true, "fr"},
		{":japanese", true, "ja"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			part, ok := p.Parse(tt.input)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.want, part.Value)
			}
		})
	}
}

func TestLanguageParser_Unknown(t *testing.T) {
	p := LanguageParser{}
	assert.True(t, p.Check(":xyz"))
	_, ok := p.Parse(":xyz")
	assert.False(t, ok)
}

func TestExternalBangParser(t *testing.T) {
	p := ExternalBangParser{}
	tests := []struct {
		input  string
		wantOK bool
		want   string
	}{
		{"!!g", true, "g"},
		{"!!ddg", true, "ddg"},
		{"!!google_images", true, "google images"},
		{"!g", false, ""},
		{"golang", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.wantOK, p.Check(tt.input))
			if tt.wantOK {
				part, ok := p.Parse(tt.input)
				assert.True(t, ok)
				assert.Equal(t, tt.want, part.Value)
			}
		})
	}
}

func TestBangParser_EngineShortcut(t *testing.T) {
	shortcuts := map[string]string{"gh": "github", "so": "stackoverflow", "g": "google"}
	names := []string{"google", "github", "stackoverflow", "wikipedia"}
	categories := []models.Category{"general", "images", "news", "videos"}

	p := BangParser{shortcuts: shortcuts, names: names, categories: categories}

	tests := []struct {
		input     string
		wantOK    bool
		wantValue string
		wantKind  string
	}{
		{"!gh", true, "github", "engine"},
		{"!so", true, "stackoverflow", "engine"},
		{"!wikipedia", true, "wikipedia", "engine"},
		{"!images", true, "images", "category"},
		{"!news", true, "news", "category"},
		{"!unknown", false, "", ""},
		{"golang", false, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.wantOK, p.Check(tt.input))
			if tt.wantOK {
				part, ok := p.Parse(tt.input)
				assert.True(t, ok)
				assert.Equal(t, tt.wantValue, part.Value)
				assert.Equal(t, tt.wantKind, part.Extra["kind"])
			}
		})
	}
}

func TestBangParser_UnknownPreserved(t *testing.T) {
	shortcuts := map[string]string{}
	names := []string{"google"}
	categories := []models.Category{"general"}

	p := BangParser{shortcuts: shortcuts, names: names, categories: categories}
	assert.False(t, p.Check("!unknown"))
}

func TestParseBangPriority(t *testing.T) {
	extP := ExternalBangParser{}
	bangP := NewBangParser(map[string]string{"g": "google"}, []string{"google"}, []models.Category{"general"})

	input := "!!g"
	assert.True(t, extP.Check(input), "ExternalBangParser must match !!g")
	assert.False(t, bangP.Check(input), "BangParser must NOT match !!g")

	input2 := "!g"
	assert.False(t, extP.Check(input2), "ExternalBangParser must NOT match !g")
	assert.True(t, bangP.Check(input2), "BangParser must match !g")
}

func TestAutocompleteTriggerParser(t *testing.T) {
	p := AutocompleteTriggerParser{}

	assert.True(t, p.Check("?", true))
	assert.False(t, p.Check("?", false))
	assert.False(t, p.Check("golang", true))

	part, ok := p.Parse("?")
	assert.True(t, ok)
	assert.Equal(t, PartAutocompleteTrigger, part.Type)
}
