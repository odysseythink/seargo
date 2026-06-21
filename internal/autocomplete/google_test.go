package autocomplete

import (
	"strings"
	"encoding/json"
	"testing"
)

func TestGoogleJSONPStripping(t *testing.T) {
	input := `)]}'` + "\n" +
		`["golang",[["result1",0,[512]],["result <em>two</em>",0,[512]]],{"a":"x"}]`

	body := input
	if idx := strings.IndexByte(body, '['); idx >= 0 {
		body = body[idx:]
	}

	var data []interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data) < 2 {
		t.Fatal("expected at least 2 elements in top-level array")
	}
	group, ok := data[1].([]interface{})
	if !ok {
		t.Fatal("expected data[1] to be an array")
	}
	if len(group) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(group))
	}

	first := group[0].([]interface{})[0].(string)
	if first != "result1" {
		t.Fatalf("expected 'result1', got %q", first)
	}

	second := group[1].([]interface{})[0].(string)
	second = htmlTagRE.ReplaceAllString(second, "")
	if second != "result two" {
		t.Fatalf("expected 'result two' after tag strip, got %q", second)
	}
}
