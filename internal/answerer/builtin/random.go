package builtin

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/seargo/seargo/internal/answerer"
	"github.com/seargo/seargo/pkg/models"
)

func init() {
	if as := answerer.GlobalAnswerer(); as != nil {
		as.Register(newRandomAnswerer())
	}
}

type randomAnswerer struct {
	info answerer.AnswererInfo
}

func newRandomAnswerer() *randomAnswerer {
	return &randomAnswerer{
		info: answerer.AnswererInfo{
			Name:        "random",
			Description: "Generate random values (string, int, float, sha256, uuid, color)",
			Keywords:    []string{"random", "rand"},
			Examples: []string{
				"random string",
				"random int",
				"random uuid",
			},
		},
	}
}

func (a *randomAnswerer) Keywords() []string {
	return a.info.Keywords
}

func (a *randomAnswerer) Info() answerer.AnswererInfo {
	return a.info
}

func (a *randomAnswerer) Answer(ctx *answerer.AnswerContext) []models.Result {
	parts := strings.Fields(ctx.Query)
	if len(parts) < 2 {
		return nil
	}
	typ := parts[1]
	var answer string
	var err error

	switch typ {
	case "string":
		answer, err = randomAlphaNumeric(16)
	case "int":
		var n int32
		n, err = randomInt32()
		answer = strconv.FormatInt(int64(n), 10)
	case "float":
		var f float64
		f, err = randomFloat64()
		answer = strconv.FormatFloat(f, 'f', 6, 64)
	case "sha256":
		var s string
		s, err = randomAlphaNumeric(16)
		if err == nil {
			h := sha256.Sum256([]byte(s))
			answer = fmt.Sprintf("%x", h)
		}
	case "uuid":
		answer, err = randomUUID()
	case "color":
		answer, err = randomColor()
	default:
		return nil
	}
	if err != nil {
		return nil
	}
	return []models.Result{{
		Kind:    "answer",
		Title:   answer,
		Content: fmt.Sprintf("Random %s generated", typ),
		Engine:  "random",
	}}
}

// randomAlphaNumeric generates a random alphanumeric string of the given length.
func randomAlphaNumeric(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[idx.Int64()]
	}
	return string(result), nil
}

// randomInt32 generates a random int32.
func randomInt32() (int32, error) {
	max := big.NewInt(1 << 31)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return int32(n.Int64()), nil
}

// randomFloat64 generates a random float64 in [0, 1).
func randomFloat64() (float64, error) {
	max := new(big.Int).Lsh(big.NewInt(1), 53) // 2^53
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return float64(n.Int64()) / float64(1<<53), nil
}

// randomUUID generates a RFC 4122 version 4 UUID.
func randomUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Set version 4
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant 10xx
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// randomColor generates a random hex color (#RRGGBB).
func randomColor() (string, error) {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("#%02x%02x%02x", b[0], b[1], b[2]), nil
}
