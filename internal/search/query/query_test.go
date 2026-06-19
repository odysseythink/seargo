package query

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
