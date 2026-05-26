package latch

// auth.go — admin auth via dock /internal/v1/auth/verify, plus the
// agent-token middleware for /api/latch/agent/* that authenticates
// against the local latch_service_node_agent_tokens table.

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	ctxKeyUserID      = "user_id"
	ctxKeyUserRole    = "user_role"
	ctxKeyWorkspaceID = "workspace_id"
)

// requireAdminViaDock extracts Bearer → Dock.AuthVerify → role=admin.
// Sets user_id / user_role / workspace_id on the gin context.
func (p *Plugin) requireAdminViaDock() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractAccessToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		res, err := p.Dock.AuthVerify(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}
		if !strings.EqualFold(res.Role, "admin") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
		c.Set(ctxKeyUserID, res.UserID)
		c.Set(ctxKeyUserRole, res.Role)
		c.Set(ctxKeyWorkspaceID, res.WorkspaceID)

		// Closed-by-default tenant access gate (Sprint 2 / task #196).
		// Root workspace always passes via dock-side bypass; non-root
		// requires an explicit workspace_plugin_access row enabled by
		// the platform admin. Fail-closed on lookup error.
		access, err := p.Dock.WorkspacePluginAccess(res.WorkspaceID, p.Name)
		if err != nil || access == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "plugin access check failed"})
			return
		}
		if !access.Enabled {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "workspace not granted access to latch"})
			return
		}
		c.Next()
	}
}

// requireAuthViaDock — same Bearer + AuthVerify pattern but does NOT
// require admin role. Used by GET /api/latch/profiles (user list).
func (p *Plugin) requireAuthViaDock() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractAccessToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		res, err := p.Dock.AuthVerify(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}
		c.Set(ctxKeyUserID, res.UserID)
		c.Set(ctxKeyUserRole, res.Role)
		c.Set(ctxKeyWorkspaceID, res.WorkspaceID)

		// Closed-by-default tenant access gate (Sprint 2 / task #196).
		// Root workspace always passes via dock-side bypass; non-root
		// requires an explicit workspace_plugin_access row enabled by
		// the platform admin. Fail-closed on lookup error.
		access, err := p.Dock.WorkspacePluginAccess(res.WorkspaceID, p.Name)
		if err != nil || access == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "plugin access check failed"})
			return
		}
		if !access.Enabled {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "workspace not granted access to latch"})
			return
		}
		c.Next()
	}
}

// requireAgentToken — authenticates /api/latch/agent/* via the locally
// stored latch_service_node_agent_tokens table. Mirrors dock's
// AgentAuthMiddleware. Sets agent_node / agent_node_id / agent_token_id
// on the gin context.
func (p *Plugin) requireAgentToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractAccessToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			c.Abort()
			return
		}
		node, tokenID, err := p.authenticateLatchAgentToken(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			c.Abort()
			return
		}
		if node == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent token"})
			c.Abort()
			return
		}
		if err := p.touchLatchAgentToken(tokenID, time.Now()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			c.Abort()
			return
		}
		c.Set("agent_node", node)
		c.Set("agent_node_id", node.ID)
		c.Set("agent_token_id", tokenID)
		c.Next()
	}
}

// extractAccessToken: Bearer header → ?access_token= → cookie. Same
// fallback chain as dock so iOS / browser clients work the same.
func extractAccessToken(c *gin.Context) string {
	if v := strings.TrimSpace(c.GetHeader("Authorization")); v != "" {
		if strings.HasPrefix(strings.ToLower(v), "bearer ") {
			return strings.TrimSpace(v[7:])
		}
	}
	if v := strings.TrimSpace(c.Query("access_token")); v != "" {
		return v
	}
	if v, err := c.Cookie("access_token"); err == nil && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return ""
}
