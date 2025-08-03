package client

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	sdkerrors "github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// ProjectContextManager provides basic project context management for Claude Code integration.
// Simplified to match official SDK scope - only provides working directory information.
type ProjectContextManager struct {
	client          *ClaudeCodeClient
	cachedContext   *types.ProjectContext
	lastCacheUpdate time.Time
	cacheDuration   time.Duration
	mu              sync.RWMutex
}

// NewProjectContextManager creates a new project context manager.
func NewProjectContextManager(client *ClaudeCodeClient) *ProjectContextManager {
	return &ProjectContextManager{
		client:        client,
		cacheDuration: 5 * time.Minute, // Cache for 5 minutes
	}
}

// GetEnhancedProjectContext returns basic project context.
// Simplified to match official SDK scope - no complex analysis.
func (pm *ProjectContextManager) GetEnhancedProjectContext(ctx context.Context) (*types.ProjectContext, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if cached context is still valid
	if pm.cachedContext != nil && time.Since(pm.lastCacheUpdate) < pm.cacheDuration {
		return pm.cachedContext, nil
	}

	// Get base context from client
	baseContext, err := pm.client.GetProjectContext(ctx)
	if err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "BASE_CONTEXT", "failed to get base project context")
	}

	// Ensure working directory is absolute
	if baseContext.WorkingDirectory != "" {
		if absPath, err := filepath.Abs(baseContext.WorkingDirectory); err == nil {
			baseContext.WorkingDirectory = absPath
		}
	}

	// Cache the context
	pm.cachedContext = baseContext
	pm.lastCacheUpdate = time.Now()

	return baseContext, nil
}

// InvalidateCache invalidates the cached project context.
func (pm *ProjectContextManager) InvalidateCache() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.cachedContext = nil
	pm.lastCacheUpdate = time.Time{}
}

// SetCacheDuration sets the cache duration for project context.
func (pm *ProjectContextManager) SetCacheDuration(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.cacheDuration = duration
}

// GetCacheInfo returns information about the cache status.
func (pm *ProjectContextManager) GetCacheInfo() map[string]any {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	info := map[string]any{
		"cache_duration": pm.cacheDuration.String(),
		"is_cached":      pm.cachedContext != nil,
		"cache_age":      time.Since(pm.lastCacheUpdate).String(),
	}

	if pm.cachedContext != nil {
		info["last_update"] = pm.lastCacheUpdate.Format(time.RFC3339)
		info["cache_valid"] = time.Since(pm.lastCacheUpdate) < pm.cacheDuration
	}

	return info
}
