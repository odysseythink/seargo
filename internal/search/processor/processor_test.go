package processor

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockSuspension struct {
	banned map[string]bool
}

func newMockSuspension() *mockSuspension {
	return &mockSuspension{banned: make(map[string]bool)}
}

func (m *mockSuspension) Ban(engineName, errorClass string) {
	m.banned[engineName] = true
}

func (m *mockSuspension) IsSuspended(engineName string) bool {
	return m.banned[engineName]
}

func TestBaseProcessor_RecordResultSuccess(t *testing.T) {
	ms := newMockSuspension()
	bp := &BaseProcessor{engineName: "test", suspension: ms}

	bp.RecordResult(true, nil)
	assert.False(t, ms.IsSuspended("test"), "success should not suspend")
}

func TestBaseProcessor_RecordResultFailure(t *testing.T) {
	ms := newMockSuspension()
	bp := &BaseProcessor{engineName: "test", suspension: ms}

	bp.RecordResult(false, errors.New("403 access denied"))
	assert.True(t, ms.IsSuspended("test"), "failure should suspend")
}

func TestBaseProcessor_Suspended(t *testing.T) {
	ms := newMockSuspension()
	bp := &BaseProcessor{engineName: "test", suspension: ms}

	assert.False(t, bp.Suspended())
	ms.Ban("test", "SearxEngineCaptcha")
	assert.True(t, bp.Suspended())
}

func TestBaseProcessor_RecordResultNilSuspension(t *testing.T) {
	bp := &BaseProcessor{engineName: "test", suspension: nil}
	bp.RecordResult(false, errors.New("err"))
	assert.False(t, bp.Suspended())
}
