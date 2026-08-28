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
		Description: "Daily services and shortcuts", Icon: "mdi:home-variant",
		Layout: models.LayoutRows, Width: models.WidthDefault, Privacy: models.PrivacyPublic,
		IsMain: true, IsDefault: true,
	}
	media := &models.Dashboard{
		ID: uuid.NewString(), OwnerID: adminID, Name: "Media", Slug: "media",
		Description: "Streaming, books, music, and downloads", Icon: "mdi:multimedia",
		Layout: models.LayoutMasonry, Width: models.WidthWide, Privacy: models.PrivacyUsers,
	}
	knowledge := &models.Dashboard{
		ID: uuid.NewString(), OwnerID: adminID, Name: "Knowledge", Slug: "knowledge",
		Description: "Documentation, learning, and reference material", Icon: "mdi:bookshelf",
		Layout: models.LayoutRows, Width: models.WidthDefault, Privacy: models.PrivacyPublic, CleanMode: true,
	}
	dev := &models.Dashboard{
		ID: uuid.NewString(), OwnerID: userID, Name: "Dev Tools", Slug: "dev",
		Description: "Development tools and project resources", Icon: "mdi:code-braces",
		Layout: models.LayoutColumns, Width: models.WidthDefault, Privacy: models.PrivacyPrivate,
	}
	work := &models.Dashboard{
		ID: uuid.NewString(), OwnerID: userID, Name: "Work", Slug: "work",
		Description: "Communication, planning, and productivity", Icon: "mdi:briefcase-outline",
		Layout: models.LayoutRows, Width: models.WidthWide, Privacy: models.PrivacyUsers, IsDefault: true,
	}
	homelab := &models.Dashboard{
		ID: uuid.NewString(), OwnerID: userID, Name: "Home Lab", Slug: "homelab",
		Description: "Local infrastructure and smart home", Icon: "mdi:home-automation",
		Layout: models.LayoutMasonry, Width: models.WidthWide, Privacy: models.PrivacyPrivate, CleanMode: true,
	}
	for _, d := range []*models.Dashboard{home, media, knowledge, dev, work, homelab} {
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
		{title: "Monitoring", desc: "Health and availability", icon: "mdi:chart-line", size: models.Size1x2, items: []itemSpec{
			{"Uptime Kuma", "Service status", "https://uptime.kuma.pet", "simple-icons:uptimekuma"},
			{"Prometheus", "Metrics database", "https://prometheus.io", "simple-icons:prometheus"},
			{"Loki", "Log aggregation", "https://grafana.com/oss/loki", "simple-icons:grafana"},
		}},
		{title: "Cloud", desc: "External infrastructure", icon: "mdi:cloud-outline", size: models.Size1x1, items: []itemSpec{
			{"Cloudflare", "", "https://dash.cloudflare.com", "simple-icons:cloudflare"},
			{"DigitalOcean", "", "https://cloud.digitalocean.com", "simple-icons:digitalocean"},
			{"AWS", "", "https://console.aws.amazon.com", "mdi:aws"},
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
		{title: "Music", desc: "Listen and discover", icon: "mdi:music", size: models.Size1x1, items: []itemSpec{
			{"Spotify", "", "https://open.spotify.com", "mdi:spotify"},
			{"YouTube Music", "", "https://music.youtube.com", "mdi:youtube"},
			{"Last.fm", "", "https://last.fm", "simple-icons:lastdotfm"},
		}},
		{title: "Reading", desc: "Books, articles, and feeds", icon: "mdi:book-open-page-variant", size: models.Size1x2, items: []itemSpec{
			{"Kavita", "Digital library", "https://www.kavitareader.com", "mdi:bookshelf"},
			{"FreshRSS", "Feed reader", "https://freshrss.org", "mdi:rss"},
			{"Pocket", "Read it later", "https://getpocket.com", "simple-icons:pocket"},
		}},
	}); err != nil {
		log.Fatal(err)
	}

	if err := seedGroups(knowledge.ID, []groupSpec{
		{title: "Documentation", desc: "Technical references", icon: "mdi:file-document-multiple-outline", size: models.Size1x2, items: []itemSpec{
			{"MDN", "Web platform documentation", "https://developer.mozilla.org", "simple-icons:mdnwebdocs"},
			{"Go", "Go language documentation", "https://pkg.go.dev", "simple-icons:go"},
			{"Svelte", "Framework documentation", "https://svelte.dev/docs", "simple-icons:svelte"},
			{"Tailwind CSS", "Utility CSS reference", "https://tailwindcss.com/docs", "simple-icons:tailwindcss"},
		}},
		{title: "Learning", desc: "Courses and tutorials", icon: "mdi:school-outline", size: models.Size1x1, items: []itemSpec{
			{"freeCodeCamp", "", "https://freecodecamp.org", "simple-icons:freecodecamp"},
			{"Exercism", "", "https://exercism.org", "simple-icons:exercism"},
			{"YouTube", "", "https://youtube.com", "mdi:youtube"},
			{"Wikipedia", "", "https://wikipedia.org", "mdi:wikipedia"},
		}},
		{title: "Communities", desc: "Questions and discussions", icon: "mdi:account-group-outline", size: models.Size1x2, items: []itemSpec{
			{"Stack Overflow", "Programming questions", "https://stackoverflow.com", "mdi:stack-overflow"},
			{"DEV Community", "Developer articles", "https://dev.to", "simple-icons:devdotto"},
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
		{title: "Repositories", desc: "Source code hosting", icon: "mdi:source-repository-multiple", size: models.Size1x2, items: []itemSpec{
			{"GitHub", "Personal repositories", "https://github.com", "mdi:github"},
			{"GitLab", "Team repositories", "https://gitlab.com", "mdi:gitlab"},
			{"Codeberg", "Open-source projects", "https://codeberg.org", "mdi:source-repository"},
		}},
		{title: "Packages", desc: "Package registries", icon: "mdi:package-variant-closed", size: models.Size1x1, items: []itemSpec{
			{"npm", "", "https://npmjs.com", "mdi:npm"},
			{"Go Packages", "", "https://pkg.go.dev", "simple-icons:go"},
			{"Docker Hub", "", "https://hub.docker.com", "mdi:docker"},
			{"GitHub Packages", "", "https://github.com/features/packages", "mdi:package-variant"},
		}},
	}); err != nil {
		log.Fatal(err)
	}

	if err := seedGroups(work.ID, []groupSpec{
		{title: "Communication", icon: "mdi:message-text-outline", size: models.Size1x1, items: []itemSpec{
			{"Slack", "", "https://slack.com", "mdi:slack"},
			{"Teams", "", "https://teams.microsoft.com", "mdi:microsoft-teams"},
			{"Gmail", "", "https://mail.google.com", "mdi:gmail"},
			{"Meet", "", "https://meet.google.com", "mdi:google-meet"},
		}},
		{title: "Planning", desc: "Projects and tasks", icon: "mdi:clipboard-check-outline", size: models.Size1x2, items: []itemSpec{
			{"Linear", "Issue tracking", "https://linear.app", "simple-icons:linear"},
			{"Notion", "Team workspace", "https://notion.so", "mdi:notion"},
			{"Google Calendar", "Meetings and events", "https://calendar.google.com", "mdi:calendar"},
		}},
		{title: "Files", desc: "Documents and storage", icon: "mdi:folder-outline", size: models.Size1x2, items: []itemSpec{
			{"Google Drive", "Shared documents", "https://drive.google.com", "mdi:google-drive"},
			{"Dropbox", "File storage", "https://dropbox.com", "mdi:dropbox"},
		}},
	}); err != nil {
		log.Fatal(err)
	}

	if err := seedGroups(homelab.ID, []groupSpec{
		{title: "Servers", desc: "Compute and containers", icon: "mdi:server-network", size: models.Size1x2, items: []itemSpec{
			{"Proxmox", "Virtualization", "https://proxmox.com", "simple-icons:proxmox"},
			{"Portainer", "Container management", "https://portainer.io", "simple-icons:portainer"},
			{"TrueNAS", "Network storage", "https://truenas.com", "simple-icons:truenas"},
		}},
		{title: "Smart home", desc: "Automation and devices", icon: "mdi:home-automation", size: models.Size1x1, items: []itemSpec{
			{"Home Assistant", "", "https://home-assistant.io", "mdi:home-assistant"},
			{"Zigbee2MQTT", "", "https://zigbee2mqtt.io", "simple-icons:zigbee2mqtt"},
			{"Node-RED", "", "https://nodered.org", "simple-icons:nodered"},
		}},
		{title: "Network", desc: "Routing and DNS", icon: "mdi:lan", size: models.Size1x2, items: []itemSpec{
			{"OpenWrt", "Router administration", "https://openwrt.org", "simple-icons:openwrt"},
			{"Pi-hole", "Network-wide ad blocking", "https://pi-hole.net", "simple-icons:pihole"},
			{"Tailscale", "Private mesh network", "https://tailscale.com", "simple-icons:tailscale"},
		}},
	}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Seed complete.")
	fmt.Println("  admin / admin123  — Home (main), Media, Knowledge")
	fmt.Println("  demo  / user123   — Dev Tools, Work (default), Home Lab")
}
