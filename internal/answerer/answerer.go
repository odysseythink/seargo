package answerer

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/seargo/seargo/pkg/models"
)

// AnswererInfo describes an answerer for display and keyword matching.
type AnswererInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Examples    []string `json:"examples,omitempty"`
	Keywords    []string `json:"keywords"`
}

// AnswerContext carries the context for an answerer invocation.
type AnswerContext struct {
	Query       string
	Locale      string
	Preferences map[string]any
}

// Answerer is the interface for instant-answer providers.
type Answerer interface {
	Keywords() []string
	Info() AnswererInfo
	Answer(ctx *AnswerContext) []models.Result
}

// AnswererStorage manages answerer registration and keyword dispatch.
// It maintains a keyword -> []Answerer index for O(1) lookup.
type AnswererStorage struct {
	mu        sync.RWMutex
	answerers []Answerer
	keywordTo map[string][]Answerer // lowercase keyword -> answerers
}

// NewAnswererStorage creates an empty answerer storage.
func NewAnswererStorage() *AnswererStorage {
	return &AnswererStorage{
		keywordTo: make(map[string][]Answerer),
	}
}

var (
	ansMu         sync.RWMutex
	globalAnswerer *AnswererStorage
)

// SetGlobalAnswerer sets the global answerer storage.
func SetGlobalAnswerer(as *AnswererStorage) {
	ansMu.Lock()
	defer ansMu.Unlock()
	globalAnswerer = as
}

// GlobalAnswerer returns the global answerer storage.
func GlobalAnswerer() *AnswererStorage {
	ansMu.RLock()
	defer ansMu.RUnlock()
	return globalAnswerer
}

// ResetForTest clears global state (tests only).
func ResetForTest() {
	ansMu.Lock()
	defer ansMu.Unlock()
	globalAnswerer = nil
}

// Register adds an answerer and indexes it by all its keywords.
func (as *AnswererStorage) Register(a Answerer) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.answerers = append(as.answerers, a)
	for _, kw := range a.Keywords() {
		lower := strings.ToLower(kw)
		as.keywordTo[lower] = append(as.keywordTo[lower], a)
	}
}

// All returns all registered answerers.
func (as *AnswererStorage) All() []Answerer {
	as.mu.RLock()
	defer as.mu.RUnlock()
	result := make([]Answerer, len(as.answerers))
	copy(result, as.answerers)
	return result
}

// Ask dispatches to answerers whose keywords match the first word of the query.
// Returns collected results from all matching answerers.
func (as *AnswererStorage) Ask(ctx *AnswerContext) []models.Result {
	firstWord := firstWordOf(ctx.Query)
	if firstWord == "" {
		return nil
	}

	as.mu.RLock()
	answerers := as.keywordTo[firstWord]
	as.mu.RUnlock()

	if len(answerers) == 0 {
		return nil
	}

	var all []models.Result
	for _, a := range answerers {
		results := callAnswer(a, ctx)
		all = append(all, results...)
	}
	return all
}

// callAnswer wraps Answerer.Answer with panic recovery.
func callAnswer(a Answerer, ctx *AnswerContext) (results []models.Result) {
	defer func() {
		if r := recover(); r != nil {
			logAnswererPanic(a.Info().Name, r)
			results = nil
		}
	}()
	return a.Answer(ctx)
}

// --- helpers ---

func firstWordOf(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, ' '); idx > 0 {
		return strings.ToLower(s[:idx])
	}
	return strings.ToLower(s)
}

func logAnswererPanic(name string, recovered any) {
	fmt.Fprintf(os.Stderr, "[seargo] panic in answerer %s: %v\n", name, recovered)
}
