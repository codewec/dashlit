package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/bookmarks-dashboard/backend/internal/auth"
	"github.com/bookmarks-dashboard/backend/internal/config"
	"github.com/bookmarks-dashboard/backend/internal/db"
	"github.com/bookmarks-dashboard/backend/internal/handlers"
	"github.com/bookmarks-dashboard/backend/internal/legacy"
	"github.com/bookmarks-dashboard/backend/internal/updatecheck"
)

var (
	version = "dev"
	commit  = "unknown"
)

//go:embed all:static
var staticFS embed.FS

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Printf("DashLit %s (%s)\n", version, commit)
		return
	}
	cfg := config.Load()
	if cfg.VersionOverride != "" {
		version = cfg.VersionOverride
	}
	log.Printf("DashLit %s (%s)", version, commit)
	if cfg.OIDCEnabled() && cfg.OIDCInsecureSkipTLSVerify {
		log.Printf("WARNING: OIDC TLS certificate verification is disabled")
	}
	if err := cfg.ValidateStorage(); err != nil {
		serveStorageError(cfg, err)
		return
	}
	if (strings.TrimSpace(cfg.OIDCIssuer) == "") != (strings.TrimSpace(cfg.OIDCClientID) == "") {
		log.Fatal("OIDC_ISSUER and OIDC_CLIENT_ID must either both be set or both be empty")
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()

	authSvc := auth.NewService(database, cfg)
	legacyMigrator, err := legacy.NewMigrator(context.Background(), database, cfg)
	if err != nil {
		log.Printf("legacy dashboard migration disabled: %v", err)
		legacyMigrator = &legacy.Migrator{}
	}
	oidcAuthenticator, err := auth.NewOIDCAuthenticator(context.Background(), cfg)
	if err != nil {
		log.Fatalf("oidc: %v", err)
	}
	authH := handlers.NewAuthHandler(authSvc, cfg, oidcAuthenticator, legacyMigrator)
	adminH := handlers.NewAdminHandler(database, authSvc, cfg)
	dashH := handlers.NewDashboardHandler(database)
	giH := handlers.NewGroupItemHandler(database)
	iconH := handlers.NewIconHandler(database, cfg)
	systemH := handlers.NewSystemHandler(updatecheck.New(version, commit, cfg.UpdateCheckEnabled, cfg.LatestVersionOverride))

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:8080", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(auth.Middleware(authSvc))

	r.Route("/api", func(r chi.Router) {
		r.Get("/system/version", systemH.Version)
		r.Get("/auth/config", authH.Configuration)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/register", authH.Register)
		r.Get("/auth/oidc/login", authH.OIDCLogin)
		r.Get("/auth/oidc/callback", authH.OIDCCallback)
		r.Post("/auth/logout", authH.Logout)
		r.Get("/auth/me", authH.Me)
		r.With(auth.RequireAuth).Put("/auth/profile", authH.UpdateProfile)

		r.Get("/dashboards", dashH.List)
		r.Get("/dashboards/main", dashH.GetMain)
		r.Get("/dashboards/{id}", dashH.Get)
		r.With(auth.RequireAuth).Post("/dashboards", dashH.Create)
		r.With(auth.RequireAuth).Put("/dashboards/{id}", dashH.Update)
		r.With(auth.RequireAuth).Delete("/dashboards/{id}", dashH.Delete)
		r.With(auth.RequireAuth).Post("/dashboards/{id}/set-main", dashH.SetMain)
		r.With(auth.RequireAuth).Post("/dashboards/{id}/set-default", dashH.SetDefault)
		r.With(auth.RequireAuth).Post("/dashboards/{id}/clone", dashH.Clone)
		r.With(auth.RequireAuth).Get("/dashboards/{id}/export", dashH.Export)
		r.With(auth.RequireAuth).Post("/dashboards/import", dashH.Import)

		r.With(auth.RequireAuth).Post("/dashboards/{dashboardID}/groups", giH.CreateGroup)
		r.With(auth.RequireAuth).Put("/groups/{id}", giH.UpdateGroup)
		r.With(auth.RequireAuth).Delete("/groups/{id}", giH.DeleteGroup)
		r.With(auth.RequireAuth).Post("/groups/{id}/clone", giH.CloneGroup)
		r.With(auth.RequireAuth).Post("/groups/{groupID}/items", giH.CreateItem)
		r.With(auth.RequireAuth).Put("/items/{id}", giH.UpdateItem)
		r.With(auth.RequireAuth).Delete("/items/{id}", giH.DeleteItem)
		r.With(auth.RequireAuth).Post("/items/{id}/clone", giH.CloneItem)
		r.Get("/items/{id}/ping", giH.PingItem)
		r.With(auth.RequireAuth).Put("/dashboards/{dashboardID}/layout", giH.UpdateLayout)

		r.With(auth.RequireAuth).Post("/icons/upload", iconH.Upload)
		r.Get("/icons/{id}", iconH.Serve)
		r.Get("/icons/search/iconify", iconH.SearchIconify)
		r.Get("/icons/search/selfhst", iconH.SearchSelfhst)
		r.Get("/icons/iconify/{prefix}/{name}", iconH.ProxyIconify)
		r.Get("/icons/selfhst/*", iconH.ProxySelfhst)

		r.Route("/admin", func(r chi.Router) {
			r.Use(auth.RequireAdmin)
			r.Get("/overview", adminH.Overview)
			r.Put("/users/{id}", adminH.UpdateUser)
			r.Delete("/users/{id}", adminH.DeleteUser)
		})
	})

	// Static frontend
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		// Dev fallback: no embedded static
		log.Println("no embedded static, serving API only (run frontend separately or build with embed)")
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "frontend not embedded; build with make build", http.StatusNotFound)
		})
	} else {
		fileServer := http.FileServer(http.FS(static))
		r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
			path := strings.TrimPrefix(req.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			if _, err := fs.Stat(static, path); err != nil {
				// SPA fallback
				req.URL.Path = "/"
				fileServer.ServeHTTP(w, req)
				return
			}
			fileServer.ServeHTTP(w, req)
		})
	}

	log.Printf("listening on %s (data: %s)", cfg.Addr, cfg.DataDir)
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func serveStorageError(cfg *config.Config, storageErr error) {
	const docsURL = "https://codewec.github.io/dashlit/guide/installation#bind-mounts"
	const page = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>DashLit storage error</title>
  <style>
    :root { color-scheme: light dark; font-family: system-ui, sans-serif; }
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: #11111b; color: #cdd6f4; }
    main { box-sizing: border-box; width: min(680px, calc(100% - 32px)); padding: 32px; border: 1px solid #45475a; border-radius: 16px; background: #1e1e2e; }
    h1 { margin-top: 0; color: #f38ba8; font-size: 1.65rem; }
    code { overflow-wrap: anywhere; color: #f9e2af; }
    pre { overflow-x: auto; padding: 14px; border-radius: 8px; background: #11111b; }
    a { color: #89b4fa; }
  </style>
</head>
<body><main>
  <h1>DashLit cannot write to its data directory</h1>
  <p>The application is running as UID:GID <code>{{.UID}}:{{.GID}}</code> and cannot initialize its storage.</p>
  <p><code>{{.Error}}</code></p>
  <p>For a Docker bind mount such as <code>./data:/data</code>, the simplest fix is:</p>
  <pre><code>sudo chmod -R 777 ./data</code></pre>
  <p>For more restrictive ownership, use:</p>
  <pre><code>sudo chown -R 10001:10001 ./data
sudo chmod -R 750 ./data</code></pre>
  <p>Fix the permissions and restart the container. See the <a href="{{.DocsURL}}">bind mount documentation</a> for details.</p>
</main></body></html>`

	log.Printf("STORAGE PERMISSION ERROR: %v", storageErr)
	log.Printf("DashLit is running as UID:GID %d:%d and cannot write its required storage", os.Getuid(), os.Getgid())
	log.Printf("data directory: %s", cfg.DataDir)
	log.Printf("for a Docker ./data bind mount, run: sudo chmod -R 777 ./data")
	log.Printf("documentation: %s", docsURL)

	tmpl := template.Must(template.New("storage-error").Parse(page))
	data := struct {
		UID     int
		GID     int
		Error   string
		DocsURL string
	}{os.Getuid(), os.Getgid(), storageErr.Error(), docsURL}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusInternalServerError)
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("render storage error page: %v", err)
		}
	})

	log.Printf("serving the storage error page on %s; fix the permissions and restart DashLit", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		log.Printf("storage error page server: %v", err)
		select {}
	}
}
