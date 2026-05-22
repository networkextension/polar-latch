// Command latch-svc is the latch (proxy + rule + profile) plugin binary.
// Phase 2-W2 skeleton — DB pool + dock heartbeat + /healthz + moved handlers.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/networkextension/polar-latch/internal/latch"
)

func main() {
	cfg := latch.Config{
		DBDSN:        envOrDefault("POLAR_LATCH_DB_DSN", "postgres://ideamesh:test123456@127.0.0.1:5432/polar_latch?sslmode=disable"),
		DockBase:     envOrDefault("POLAR_DOCK_URL", "http://127.0.0.1:8080"),
		PluginName:   envOrDefault("POLAR_PLUGIN_NAME", "latch"),
		PluginToken:  os.Getenv("POLAR_PLUGIN_TOKEN"),
		Listen:       envOrDefault("POLAR_LATCH_LISTEN", "127.0.0.1:8098"),
		BuildVersion: envOrDefault("POLAR_LATCH_VERSION", "0.0.1"),
		BlobDir:      envOrDefault("POLAR_LATCH_BLOB_DIR", "/Users/local/latch-svc-data"),
		MetricsToken: os.Getenv("POLAR_LATCH_METRICS_TOKEN"),
	}
	if strings.TrimSpace(cfg.PluginToken) == "" {
		log.Fatal("POLAR_PLUGIN_TOKEN unset — get plaintext from /admin-plugins.html")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	plugin, err := latch.New(ctx, cfg)
	if err != nil {
		log.Fatalf("latch.New: %v", err)
	}
	defer plugin.Close()

	gin.SetMode(envOrDefault("GIN_MODE", gin.ReleaseMode))
	r := gin.New()
	r.Use(gin.Recovery())
	plugin.RegisterRoutes(r)
	plugin.Start(ctx)

	srv := &http.Server{Addr: cfg.Listen, Handler: r, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Printf("latch-svc listening on %s (dock=%s, name=%s, ver=%s, blob=%s)",
			cfg.Listen, cfg.DockBase, cfg.PluginName, cfg.BuildVersion, cfg.BlobDir)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	<-ctx.Done()
	log.Print("latch-svc: shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("latch-svc: shutdown: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
