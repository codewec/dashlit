<h1 align="center">DashLit</h1>
<p align="center">
    <i>DashLit is a simple, self-hosted Startpage solution. It’s incredibly easy to set up and use, and its built-in editors let you quickly create your own application hub – even with a convenient drag-and-drop interface. You don’t even need to edit any files!</i>
    <br/><br/>
    <img width="130" alt="DashLit" src="https://raw.githubusercontent.com/codewec/dashlit/main/static/favicon.svg"/>
    <br/> <br/>
    <img src="https://img.shields.io/github/v/release/codewec/dashlit?logo=hackthebox&color=609966&logoColor=fff" alt="Current Version"/>
    <img src="https://img.shields.io/github/last-commit/codewec/dashlit?logo=github&color=609966&logoColor=fff" alt="Last commit"/>
    <a href="https://github.com/codewec/dashlit/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-609966?logo=opensourceinitiative&logoColor=fff" alt="License MIT"/></a>
    <a href="https://dashlit.cwec.dev/" target="_blank"><img src="https://img.shields.io/badge/live-demo-609966"/></a>
    <br/><br/>
    <img src="https://raw.githubusercontent.com/codewec/dashlit/main/.github/main_page.png" alt="DashLit" width="100%"/>
</p>

## 🚀 Getting started

### Docker

This Docker image is published on the GitHub container registry - `ghcr.io/codewec/dashlit`.

#### Minimal configuration without password

```yaml
services:
  app:
    container_name: dashlit-app
    image: ghcr.io/codewec/dashlit:latest
    restart: unless-stopped
    environment:
      ORIGIN: '${ORIGIN:-http://localhost:3000}' # please provide URL if different
    ports:
      - '3000:3000'
    volumes:
      - ./data:/app/data
```

#### Full configuration with password

```yaml
services:
  app:
    container_name: dashlit-app
    image: ghcr.io/codewec/dashlit:latest
    environment:
      ORIGIN: '${ORIGIN:-http://localhost:3000}' # please provide URL if different
      NODE_ENV: '${NODE_ENV:-production}' # optional for production environment
      HOST_HEADER: '${HOST_HEADER:-HOST}' # optional for nginx reverse proxy
      ADDRESS_HEADER: '${ADDRESS_HEADER:-X-Real-IP}' # optional for nginx reverse proxy
      PROTOCOL_HEADER: '${PROTOCOL_HEADER:-X-Forwarded-Proto}' # optional for nginx reverse proxy
      PASSWORD: '${PASSWORD:-password}'
      SECRET_KEY: '${SECRET_KEY:-any-secret-string-for-jwt-auth}' # optional key for JWT authentication
    restart: unless-stopped
    ports:
      - '3000:3000'
    volumes:
      - ./data:/app/data
```
