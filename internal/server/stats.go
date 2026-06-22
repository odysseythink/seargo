package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/seargo/seargo/internal/metrics"
)

// handleStatsEngines returns per-engine reliability, timing, and error stats.
func (s *Server) handleStatsEngines(c *gin.Context) {
	engines := s.enginesStatsStore.SnapshotAll()
	if engines == nil {
		engines = []*metrics.EngineSnapshot{}
	}
	c.JSON(http.StatusOK, gin.H{"engines": engines})
}

// handleStatsErrors returns per-engine error counts by class.
func (s *Server) handleStatsErrors(c *gin.Context) {
	allSnapshots := s.enginesStatsStore.SnapshotAll()
	type errorEntry struct {
		Engine      string            `json:"engine"`
		TotalErrors uint64            `json:"total_errors"`
		ByClass     map[string]uint64 `json:"by_class"`
	}
	result := make([]errorEntry, 0)
	for _, snap := range allSnapshots {
		if len(snap.ErrorCounts) == 0 {
			continue
		}
		var total uint64
		for _, c := range snap.ErrorCounts {
			total += c
		}
		result = append(result, errorEntry{
			Engine:      snap.Engine,
			TotalErrors: total,
			ByClass:     snap.ErrorCounts,
		})
	}
	if result == nil {
		result = []errorEntry{}
	}
	c.JSON(http.StatusOK, gin.H{"errors": result})
}
