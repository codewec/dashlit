---
layout: home

hero:
  name: DashLit
  text: Your services, one clear dashboard
  tagline: A fast, modern, self-hosted home for links, tools, and service status — packaged as a single container.
  image:
    src: /logo.svg
    alt: DashLit logo
  actions:
    - theme: brand
      text: Install DashLit
      link: /guide/installation
    - theme: alt
      text: Explore the guide
      link: /guide/getting-started
    - theme: alt
      text: Join the discussion
      link: https://github.com/codewec/dashlit/discussions

features:
  - icon: 🧭
    title: Multiple dashboards
    details: Organize links into dashboards and groups, then choose rows, columns, or masonry layouts.
  - icon: 🔐
    title: Access on your terms
    details: Use passwords, OIDC, or both. Make dashboards public, available to signed-in users, or private.
  - icon: 📡
    title: Service awareness
    details: Check linked services and show live availability directly alongside your shortcuts.
  - icon: 🎨
    title: Made to feel at home
    details: Choose light and dark themes, compact clean mode, custom icons, and wide layouts.
  - icon: ✨
    title: Two icon libraries
    details: Search selfh.st/icons and Iconify together, with fast independent results and thousands of service and general-purpose icons.
  - icon: 🌓
    title: Theme-aware icons
    details: DashLit pairs light and dark selfh.st variants automatically and keeps monochrome Iconify icons visible on dark backgrounds.
  - icon: 📦
    title: Simple to operate
    details: A compact standalone binary with no runtime dependencies, also shipped as one container with SQLite-backed persistent storage.
  - icon: 🔁
    title: Easy to move
    details: Import, export, and clone dashboards — including assisted migration from legacy DashLit releases.
---

## A useful start page, without the overhead

DashLit gives teams, homelabs, and individuals a focused place to reach the services they use every day. It ships as a Go server with the Svelte interface embedded, so deployment is just one container and one persistent data directory.

<div class="screenshot-placeholder">
  <div><strong>Hero screenshot placeholder</strong><br>Add a wide dashboard screenshot here later.</div>
</div>

## Ready in a few seconds

Create `docker-compose.yml`:

```yaml
services:
  dashlit:
    image: ghcr.io/codewec/dashlit:beta
    ports:
      - '3000:8080'
    environment:
      JWT_SECRET: replace-with-a-long-random-secret
    volumes:
      - dashlit-data:/data

volumes:
  dashlit-data:
```

Start it:

```bash
docker compose up -d
```

Open `http://localhost:3000`, create the first account, and begin building your dashboard. The first account is granted administrator access automatically.

[Read the installation guide →](/guide/installation)

## Help shape DashLit

Have a setup question or want to show how you use DashLit? [Start a discussion](https://github.com/codewec/dashlit/discussions). If you found a bug or have a concrete feature proposal, [open an issue](https://github.com/codewec/dashlit/issues).
