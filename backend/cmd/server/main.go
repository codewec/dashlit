package main

import (
	"context"
	"embed"
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
)

//go:embed all:static
var staticFS embed.FS

func main() {
	cfg := config.Load()
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
		r.Get("/icons/iconify/{prefix}/{name}", iconH.ProxyIconify)

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
