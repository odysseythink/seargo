package deps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddAndContains(t *testing.T) {
	bl := NewAhmiaBlacklist()
	bl.Add("abc123")
	assert.True(t, bl.Contains("abc123"))
	assert.False(t, bl.Contains("def456"))
}

func TestEmpty(t *testing.T) {
	bl := NewAhmiaBlacklist()
	assert.False(t, bl.Contains("anything"))
}

func TestDuplicateAdd(t *testing.T) {
	bl := NewAhmiaBlacklist()
	bl.Add("abc123")
	bl.Add("abc123") // should not panic
	assert.True(t, bl.Contains("abc123"))
}

func TestLoadFromHashes(t *testing.T) {
	bl := NewAhmiaBlacklist()
	bl.LoadFromHashes([]string{"aaa", "bbb", "ccc"})
	assert.True(t, bl.Contains("aaa"))
	assert.True(t, bl.Contains("bbb"))
	assert.True(t, bl.Contains("ccc"))
	assert.False(t, bl.Contains("ddd"))
}

func TestConcurrentSafety(t *testing.T) {
	bl := NewAhmiaBlacklist()
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			bl.Add("hash")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			bl.Contains("hash")
		}
		done <- true
	}()

	<-done
	<-done
	assert.True(t, bl.Contains("hash"))
}
