package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/bookmarks-dashboard/backend/internal/config"
	"github.com/bookmarks-dashboard/backend/internal/db"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

func main() {
	cfg := config.Load()
	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	ctx := context.Background()

	for _, t := range []string{"items", "groups", "dashboards", "uploaded_icons", "users"} {
		if _, err := database.ExecContext(ctx, "DELETE FROM "+t); err != nil {
			log.Printf("warn delete %s: %v", t, err)
		}
	}

	hash := func(p string) *string {
		b, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
		s := string(b)
		return &s
	}

	adminID := uuid.NewString()
	userID := uuid.NewString()

	admin := &models.User{ID: adminID, Username: "admin", PasswordHash: hash("admin123"), Role: models.RoleAdmin}
	demo := &models.User{ID: userID, Username: "demo", PasswordHash: hash("user123"), Role: models.RoleUser}
	for _, u := range []*models.User{admin, demo} {
		if _, err := database.NewInsert().Model(u).Exec(ctx); err != nil {
			log.Fatal(err)
		}
	}

	home := &models.Dashboard{
		ID: uuid.NewString(), OwnerID: adminID, Name: "Home", Slug: "home",
		Layout: models.LayoutRows, Width: models.WidthDefault, Privacy: models.PrivacyPublic,
		IsMain: true,
	}
	media := &models.Dashboard{
		ID: uuid.NewString(), OwnerID: adminID, Name: "Media", Slug: "media",
		Layout: models.LayoutMasonry, Width: models.WidthWide, Privacy: models.PrivacyUsers,
	}
	dev := &models.Dashboard{
		ID: uuid.NewString(), OwnerID: userID, Name: "Dev Tools", Slug: "dev",
		Layout: models.LayoutColumns, Width: models.WidthDefault, Privacy: models.PrivacyPrivate,
	}
	for _, d := range []*models.Dashboard{home, media, dev} {
		if _, err := database.NewInsert().Model(d).Exec(ctx); err != nil {
			log.Fatal(err)
		}
	}

	type itemSpec struct{ title, desc, url, icon string }
	type groupSpec struct {
		title, desc, icon string
		size              models.ItemSize
		items             []itemSpec
	}

	seedGroups := func(dashID string, specs []groupSpec) error {
		for gi, gs := range specs {
			g := &models.Group{
				ID: uuid.NewString(), DashboardID: dashID,
				Title: gs.title, Description: gs.desc, Icon: gs.icon,
				ItemSize: gs.size, Position: gi,
			}
			if _, err := database.NewInsert().Model(g).Exec(ctx); err != nil {
				return err
			}
			for ii, is := range gs.items {
				it := &models.Item{
					ID: uuid.NewString(), GroupID: g.ID,
					Title: is.title, Description: is.desc, URL: is.url, Icon: is.icon,
					Position: ii,
				}
				if _, err := database.NewInsert().Model(it).Exec(ctx); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := seedGroups(home.ID, []groupSpec{
		{title: "Infrastructure", desc: "Core services", icon: "mdi:server", size: models.Size1x2, items: []itemSpec{
			{"Portainer", "Docker UI", "https://portainer.io", "simple-icons:portainer"},
			{"Traefik", "Reverse proxy", "https://traefik.io", "simple-icons:traefikproxy"},
			{"Grafana", "Metrics", "https://grafana.com", "simple-icons:grafana"},
		}},
		{title: "Quick links", icon: "mdi:lightning-bolt", size: models.Size1x1, items: []itemSpec{
			{"GitHub", "", "https://github.com", "mdi:github"},
			{"Gmail", "", "https://mail.google.com", "mdi:gmail"},
			{"Drive", "", "https://drive.google.com", "mdi:google-drive"},
			{"Calendar", "", "https://calendar.google.com", "mdi:calendar"},
		}},
	}); err != nil {
		log.Fatal(err)
	}

	if err := seedGroups(media.ID, []groupSpec{
		{title: "Streaming", desc: "Media stack", icon: "mdi:play-circle", size: models.Size1x2, items: []itemSpec{
			{"Jellyfin", "Media server", "https://jellyfin.org", "simple-icons:jellyfin"},
			{"Sonarr", "TV", "https://sonarr.tv", "mdi:television-classic"},
			{"Radarr", "Movies", "https://radarr.video", "mdi:movie"},
			{"Prowlarr", "Indexers", "https://prowlarr.com", "mdi:radar"},
		}},
	}); err != nil {
		log.Fatal(err)
	}

	if err := seedGroups(dev.ID, []groupSpec{
		{title: "Tools", desc: "Everyday dev", icon: "mdi:code-braces", size: models.Size1x1, items: []itemSpec{
			{"VS Code", "", "https://vscode.dev", "mdi:microsoft-visual-studio-code"},
			{"NPM", "", "https://npmjs.com", "mdi:npm"},
			{"Go", "", "https://go.dev", "simple-icons:go"},
			{"Svelte", "", "https://svelte.dev", "simple-icons:svelte"},
		}},
	}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Seed complete.")
	fmt.Println("  admin / admin123  — Home (main), Media")
	fmt.Println("  demo  / user123   — Dev Tools")
}
