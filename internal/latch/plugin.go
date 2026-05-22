// Package latch is the proxy / rule / profile management plugin
// extracted from Polar dock. Owns latch_proxies, latch_rules,
// latch_profiles, latch_service_nodes + heartbeats + agent tokens.
//
// Phase 2-W2 skeleton: DB pool + heartbeat + /healthz + the moved
// handlers.
package latch

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/networkextension/polar-sdk"
)

type Plugin struct {
	DB         *sql.DB
	Dock       *sdk.Client
	Name       string
	Listen     string
	Ver        string
	BlobDir    string // optional — latch handlers currently don't write blobs, kept for parity
	MetricsTok string

	metrics   *latchMetrics
	startedAt time.Time
}

type Config struct {
	DBDSN        string
	DockBase     string
	PluginName   string
	PluginToken  string
	Listen       string
	BuildVersion string
	BlobDir      string
	MetricsToken string
}

func New(ctx context.Context, cfg Config) (*Plugin, error) {
	cfg.PluginName = strings.TrimSpace(cfg.PluginName)
	if cfg.PluginName == "" {
		cfg.PluginName = "latch"
	}
	if strings.TrimSpace(cfg.DBDSN) == "" {
		return nil, errors.New("latch.New: DBDSN required")
	}
	if strings.TrimSpace(cfg.DockBase) == "" {
		return nil, errors.New("latch.New: DockBase required")
	}
	if strings.TrimSpace(cfg.PluginToken) == "" {
		return nil, errors.New("latch.New: PluginToken required")
	}

	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("open polar_latch: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping polar_latch: %w", err)
	}

	dock := sdk.NewClient(cfg.DockBase, cfg.PluginName, sdk.DeriveHMACKey(cfg.PluginToken))
	resp, err := dock.Do(http.MethodGet, "/internal/v1/ping", nil)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("dock ping: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_ = db.Close()
		return nil, fmt.Errorf("dock /ping rejected: HTTP %d", resp.StatusCode)
	}

	return &Plugin{
		DB:         db,
		Dock:       dock,
		Name:       cfg.PluginName,
		Listen:     cfg.Listen,
		Ver:        cfg.BuildVersion,
		BlobDir:    cfg.BlobDir,
		MetricsTok: cfg.MetricsToken,
		metrics:    newLatchMetrics(),
		startedAt:  time.Now(),
	}, nil
}

func (p *Plugin) RegisterRoutes(r gin.IRouter) {
	r.GET("/healthz", p.handleHealthz)
	r.GET("/metrics", p.handleMetricsExposition)

	// /api/latch/* — match the dock-side route prefix so nginx can flip
	// with a single proxy_pass redirect (see scripts/nginx/latch-svc-
	// snippet.conf). Three middleware groups: admin, authed users, and
	// agent-token (for /api/latch/agent/*).
	api := r.Group("/api")
	{
		admin := api.Group("", p.requireAdminViaDock())
		{
			// Proxies CRUD + versioning
			admin.GET("/latch/proxies", p.handleLatchProxyList)
			admin.POST("/latch/proxies", p.handleLatchProxyCreate)
			admin.GET("/latch/proxies/:group_id", p.handleLatchProxyGet)
			admin.PUT("/latch/proxies/:group_id", p.handleLatchProxyUpdate)
			admin.DELETE("/latch/proxies/:group_id", p.handleLatchProxyDelete)
			admin.GET("/latch/proxies/:group_id/versions", p.handleLatchProxyVersions)
			admin.PUT("/latch/proxies/:group_id/rollback/:version", p.handleLatchProxyRollback)

			// Rules CRUD + versioning
			admin.GET("/latch/rules", p.handleLatchRuleList)
			admin.POST("/latch/rules", p.handleLatchRuleCreate)
			admin.POST("/latch/rules/upload", p.handleLatchRuleCreateUpload)
			admin.GET("/latch/rules/:group_id", p.handleLatchRuleGet)
			admin.GET("/latch/rules/:group_id/content", p.handleLatchRuleContent)
			admin.GET("/latch/rules/:group_id/versions", p.handleLatchRuleVersions)
			admin.PUT("/latch/rules/:group_id", p.handleLatchRuleUpdate)
			admin.POST("/latch/rules/:group_id/upload", p.handleLatchRuleUpload)
			admin.DELETE("/latch/rules/:group_id", p.handleLatchRuleDelete)
			admin.PUT("/latch/rules/:group_id/rollback/:version", p.handleLatchRuleRollback)

			// Profiles admin CRUD
			admin.GET("/latch/admin/profiles", p.handleLatchAdminProfileList)
			admin.POST("/latch/admin/profiles", p.handleLatchAdminProfileCreate)
			admin.GET("/latch/admin/profiles/:id", p.handleLatchAdminProfileGet)
			admin.PUT("/latch/admin/profiles/:id", p.handleLatchAdminProfileUpdate)
			admin.DELETE("/latch/admin/profiles/:id", p.handleLatchAdminProfileDelete)

			// Service nodes
			admin.GET("/latch/admin/service-nodes", p.handleLatchServiceNodeList)
			admin.POST("/latch/admin/service-nodes", p.handleLatchServiceNodeCreate)
			admin.PUT("/latch/admin/service-nodes/:id", p.handleLatchServiceNodeUpdate)
			admin.DELETE("/latch/admin/service-nodes/:id", p.handleLatchServiceNodeDelete)
			admin.POST("/latch/admin/service-nodes/:id/agent-token", p.handleLatchServiceNodeIssueAgentToken)
		}

		// Lightweight agent runtime — agent-token auth (local DB).
		agent := api.Group("", p.requireAgentToken())
		{
			agent.POST("/latch/agent/register", p.handleLatchAgentRegister)
			agent.POST("/latch/agent/heartbeat", p.handleLatchAgentHeartbeat)
		}

		// User-facing — any logged-in user can list enabled+shareable profiles.
		authed := api.Group("", p.requireAuthViaDock())
		{
			authed.GET("/latch/profiles", p.handleLatchProfileList)
		}
	}
}

func (p *Plugin) Start(ctx context.Context) {
	go p.heartbeatLoop(ctx)
}

func (p *Plugin) Close() error {
	if p.DB != nil {
		return p.DB.Close()
	}
	return nil
}

func (p *Plugin) handleHealthz(c *gin.Context) {
	dbOK := true
	if err := p.DB.PingContext(c.Request.Context()); err != nil {
		dbOK = false
	}
	status := http.StatusOK
	if !dbOK {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{
		"plugin":         p.Name,
		"version":        p.Ver,
		"uptime_seconds": int64(time.Since(p.startedAt).Seconds()),
		"db_ok":          dbOK,
		"blob_dir":       p.BlobDir,
		"go":             runtime.Version(),
	})
}

func (p *Plugin) handleMetricsExposition(c *gin.Context) {
	if p.MetricsTok == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if c.GetHeader("Authorization") != "Bearer "+p.MetricsTok {
		c.Header("WWW-Authenticate", `Bearer realm="metrics"`)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	promhttp.HandlerFor(p.metrics.registry, promhttp.HandlerOpts{}).ServeHTTP(c.Writer, c.Request)
}

func (p *Plugin) heartbeatLoop(ctx context.Context) {
	p.beat(ctx)
	t := time.NewTicker(60 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.beat(ctx)
		}
	}
}

func (p *Plugin) beat(_ context.Context) {
	err := p.Dock.Heartbeat(sdk.HeartbeatOpts{
		Version:       p.Ver,
		Endpoint:      p.Listen,
		UptimeSeconds: int64(time.Since(p.startedAt).Seconds()),
	})
	if err != nil {
		log.Printf("latch: heartbeat failed: %v", err)
	}
}

type latchMetrics struct {
	registry *prometheus.Registry
	upGauge  prometheus.Gauge
}

func newLatchMetrics() *latchMetrics {
	m := &latchMetrics{registry: prometheus.NewRegistry()}
	m.upGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "polar_latch_up",
		Help: "Always 1 while latch-svc is serving. Phase 2-W2 placeholder.",
	})
	m.registry.MustRegister(m.upGauge)
	m.upGauge.Set(1)
	return m
}
