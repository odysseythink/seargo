package answerer

import (
	"testing"

	"github.com/seargo/seargo/pkg/models"
	"github.com/stretchr/testify/assert"
)

type mockAnswerer struct {
	info    AnswererInfo
	answers []models.Result
}

func (m *mockAnswerer) Keywords() []string                        { return m.info.Keywords }
func (m *mockAnswerer) Info() AnswererInfo                        { return m.info }
func (m *mockAnswerer) Answer(ctx *AnswerContext) []models.Result { return m.answers }

func TestAnswererStorage_RegisterAndKeywordIndex(t *testing.T) {
	ResetForTest()
	as := NewAnswererStorage()

	a := &mockAnswerer{
		info:    AnswererInfo{Name: "random", Keywords: []string{"random", "rand"}},
		answers: []models.Result{{Kind: "answer", Title: "random result"}},
	}
	as.Register(a)

	results := as.Ask(&AnswerContext{Query: "random string"})
	assert.Len(t, results, 1)
	assert.Equal(t, "random result", results[0].Title)

	results2 := as.Ask(&AnswerContext{Query: "rand int"})
	assert.Len(t, results2, 1)

	results3 := as.Ask(&AnswerContext{Query: "weather Berlin"})
	assert.Empty(t, results3)
}

func TestAnswererStorage_Ask_MultipleAnswerers_SameKeyword(t *testing.T) {
	ResetForTest()
	as := NewAnswererStorage()

	as.Register(&mockAnswerer{
		info:    AnswererInfo{Name: "a1", Keywords: []string{"calc"}},
		answers: []models.Result{{Kind: "answer", Title: "from a1"}},
	})
	as.Register(&mockAnswerer{
		info:    AnswererInfo{Name: "a2", Keywords: []string{"calc"}},
		answers: []models.Result{{Kind: "answer", Title: "from a2"}},
	})

	results := as.Ask(&AnswerContext{Query: "calc 1+1"})
	assert.Len(t, results, 2)
}

func TestAnswererStorage_Ask_NoMatchingKeyword(t *testing.T) {
	ResetForTest()
	as := NewAnswererStorage()
	as.Register(&mockAnswerer{
		info: AnswererInfo{Name: "stats", Keywords: []string{"avg", "sum"}},
	})

	results := as.Ask(&AnswerContext{Query: "hello world"})
	assert.Empty(t, results)
}

func TestAnswererStorage_Ask_EmptyQuery(t *testing.T) {
	ResetForTest()
	as := NewAnswererStorage()
	as.Register(&mockAnswerer{
		info:    AnswererInfo{Name: "r", Keywords: []string{"random"}},
		answers: []models.Result{{Kind: "answer", Title: "ok"}},
	})

	results := as.Ask(&AnswerContext{Query: ""})
	assert.Empty(t, results)
}

func TestAnswerContext_Fields(t *testing.T) {
	ctx := &AnswerContext{
		Query:       "avg 1 2 3",
		Locale:      "zh-CN",
		Preferences: map[string]any{"doi_resolver": "oadoi.org"},
	}
	assert.Equal(t, "avg 1 2 3", ctx.Query)
	assert.Equal(t, "zh-CN", ctx.Locale)
	assert.Equal(t, "oadoi.org", ctx.Preferences["doi_resolver"])
}

func TestRegister_Pending(t *testing.T) {
	ResetForTest()

	a := &mockAnswerer{
		info:    AnswererInfo{Name: "pending", Keywords: []string{"pending"}},
		answers: []models.Result{{Kind: "answer", Title: "pending result"}},
	}
	Register(a)

	// Before SetGlobalAnswerer, the answerer is queued.
	assert.Nil(t, GlobalAnswerer())

	as := NewAnswererStorage()
	SetGlobalAnswerer(as)

	// After setting global, pending answerers are flushed.
	assert.Equal(t, as, GlobalAnswerer())
	results := as.Ask(&AnswerContext{Query: "pending test"})
	assert.Len(t, results, 1)
	assert.Equal(t, "pending result", results[0].Title)
}
