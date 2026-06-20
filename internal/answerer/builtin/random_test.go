package builtin

import (
	"regexp"
	"strings"
	"testing"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/stretchr/testify/assert"
)

func TestRandomAnswerer_String(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newRandomAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "random string"})
	assert.Len(t, results, 1)
	assert.Len(t, results[0].Title, 16)
	assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9]+$`), results[0].Title)
	assert.Equal(t, "random", results[0].Engine)
}

func TestRandomAnswerer_Int(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newRandomAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "random int"})
	assert.Len(t, results, 1)
	// Should be a numeric string (possibly with a leading minus)
	assert.Regexp(t, regexp.MustCompile(`^-?\d+$`), results[0].Title)
}

func TestRandomAnswerer_Float(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newRandomAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "random float"})
	assert.Len(t, results, 1)
	// Should be a decimal string like "0.123456"
	assert.Regexp(t, regexp.MustCompile(`^\d+\.\d+$`), results[0].Title)
}

func TestRandomAnswerer_SHA256(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newRandomAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "random sha256"})
	assert.Len(t, results, 1)
	// SHA-256 hex is 64 hex characters
	assert.Len(t, results[0].Title, 64)
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{64}$`), results[0].Title)
}

func TestRandomAnswerer_UUID(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newRandomAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "random uuid"})
	assert.Len(t, results, 1)
	// UUID v4 format: 8-4-4-4-12 with version digit 4 and variant digit 8/9/a/b
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`), results[0].Title)
}

func TestRandomAnswerer_Color(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newRandomAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "random color"})
	assert.Len(t, results, 1)
	// #RRGGBB format
	assert.Regexp(t, regexp.MustCompile(`^#[0-9a-f]{6}$`), results[0].Title)
}

func TestRandomAnswerer_UnknownType(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newRandomAnswerer()
	as.Register(a)

	results := as.Ask(&answerer.AnswerContext{Query: "random unknown"})
	assert.Nil(t, results)
}

func TestRandomAnswerer_Keywords(t *testing.T) {
	a := newRandomAnswerer()
	kw := a.Keywords()
	assert.ElementsMatch(t, []string{"random", "rand"}, kw)
}

func TestRandomAnswerer_Info(t *testing.T) {
	a := newRandomAnswerer()
	info := a.Info()
	assert.Equal(t, "random", info.Name)
	assert.NotEmpty(t, info.Description)
	assert.NotEmpty(t, info.Examples)
}

func TestRandomAlphaNumeric_Length(t *testing.T) {
	s, err := randomAlphaNumeric(32)
	assert.NoError(t, err)
	assert.Len(t, s, 32)
	assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9]+$`), s)
}

func TestRandomAlphaNumeric_ZeroLength(t *testing.T) {
	s, err := randomAlphaNumeric(0)
	assert.NoError(t, err)
	assert.Empty(t, s)
}

func TestRandomUUID_Version(t *testing.T) {
	u, err := randomUUID()
	assert.NoError(t, err)
	parts := strings.Split(u, "-")
	assert.Len(t, parts, 5)
	// Version digit (13th char) should be "4"
	assert.Equal(t, "4", string(parts[2][0]))
	// Variant digit (17th char) should be 8, 9, a, or b
	variant := parts[3][0]
	assert.Contains(t, "89ab", string(variant))
}

func TestRandomColor_Format(t *testing.T) {
	c, err := randomColor()
	assert.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`^#[0-9a-f]{6}$`), c)
}

func TestRandom_ShortQuery(t *testing.T) {
	as := answerer.NewAnswererStorage()
	a := newRandomAnswerer()
	as.Register(a)

	// Only keyword, no type
	results := as.Ask(&answerer.AnswerContext{Query: "random"})
	assert.Nil(t, results)
}
