package bases

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/seargo/seargo/internal/engine"
	"github.com/seargo/seargo/pkg/models"
)

func TestOpenSearchEngine_Setup_NotImplemented(t *testing.T) {
	eng := NewOpenSearchEngine("test_os", []models.Category{models.CategoryGeneral}, OpenSearchConfig{})
	ok := eng.Setup(engine.EngineInitConfig{Name: "test_os"})
	assert.False(t, ok, "OpenSearch should return false Setup (not yet implemented)")
}

func TestCommandEngine_Setup_NotImplemented(t *testing.T) {
	eng := NewCommandEngine("test_cmd", []models.Category{models.CategoryGeneral}, CommandConfig{})
	ok := eng.Setup(engine.EngineInitConfig{Name: "test_cmd"})
	assert.False(t, ok, "Command should return false Setup (not yet implemented)")
}
