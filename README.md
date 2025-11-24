# Prestige League Season Four - Discord Bot

Discord Bot für die Verwaltung der Prestige League Season Four.

## Voraussetzungen

- Go 1.21 oder höher
- Discord Bot Token
- Docker & Docker Compose (für Container Deployment)

## Installation

1. Repository klonen:
```bash
git clone <repository-url>
cd PrestigueLeagueSeasonFour
```

2. Dependencies installieren:
```bash
go mod download
```

3. Bot Token konfigurieren:
   - Erstelle einen Bot unter https://discord.com/developers/applications
   - Setze die Umgebungsvariable `DISCORD_BOT_TOKEN` oder
   - Kopiere `.env.example` zu `.env` und füge das Token ein

## Bot starten

### Lokal mit Go

```bash
# Mit Umgebungsvariable
export DISCORD_BOT_TOKEN="dein_token_hier"
go run cmd/bot/main.go

# Oder mit PowerShell
$env:DISCORD_BOT_TOKEN="dein_token_hier"
go run cmd/bot/main.go
```

### Mit Docker Compose

```bash
# .env Datei erstellen
echo "DISCORD_BOT_TOKEN=dein_token_hier" > .env

# Container starten
docker-compose up -d

# Logs anzeigen
docker-compose logs -f

# Container stoppen
docker-compose down
```

### Docker manuell

```bash
# Image bauen
docker build -t prestigeleague-bot .

# Container starten
docker run -d \
  --name prestigeleague-bot \
  -e DISCORD_BOT_TOKEN="dein_token_hier" \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config:/app/config:ro \
  prestigeleague-bot
```

## Projektstruktur

```
.
├── cmd/
│   └── bot/
│       └── main.go           # Bot Entry Point
├── internal/
│   ├── bot/
│   │   └── bot.go            # Bot Logik und Handler
│   ├── commands/             # Command Implementierungen
│   └── handlers/             # Event Handlers
├── config/
│   └── config.yaml           # Konfigurationsdatei
├── data/                     # SQLite Datenbank (wird automatisch erstellt)
├── Dockerfile                # Docker Image Definition
├── docker-compose.yml        # Docker Compose Setup
├── .dockerignore             # Docker Build Ausschlüsse
├── go.mod
└── README.md
```

## Datenbank

Der Bot verwendet SQLite als Datenbank. Die Datenbankdatei wird automatisch im `data/` Verzeichnis erstellt.
Bei Docker Deployment wird das Volume `/app/data` gemountet, um die Persistenz sicherzustellen.

## Deployment auf Server

**Für eine vollständige Schritt-für-Schritt Anleitung siehe: [SERVER_SETUP.md](SERVER_SETUP.md)**

**Schnellstart:**
```bash
# 1. Dateien auf Server (Git oder SCP)
git clone <repository-url>
cd PrestigueLeagueSeasonFour

# 2. .env Datei konfigurieren
cp .env.example .env
nano .env  # DISCORD_BOT_TOKEN und ACME_EMAIL eintragen

# 3. DNS-Einträge setzen (prestigeleague.de → Server-IP)

# 4. Container starten
docker compose up -d

# 5. Logs prüfen
docker compose logs -f
```

Die Webseite ist dann unter `https://prestigeleague.de` erreichbar mit automatischen Let's Encrypt SSL-Zertifikaten.

Detaillierte Anleitung mit Troubleshooting: **[SERVER_SETUP.md](SERVER_SETUP.md)**

## Web Interface (Flask)

Zusätzlich zum Discord Bot gibt es eine Flask Web Anwendung, die die Matchresultate und Tabellen öffentlich anzeigt.

### Installation Web Interface

```bash
cd web
pip install -r requirements.txt
```

### Web Interface starten

```bash
cd web
python app.py
```

Die Web Anwendung läuft standardmäßig auf `http://localhost:5000` und zeigt:
- Übersicht aller Divisionen
- Tabellen mit Punkten, Siegen, Niederlagen
- Alle Matches nach Spieltagen gruppiert
- JSON API Endpoints: `/api/standings/<division>` und `/api/matches/<division>`

### Docker Compose mit Web Interface

Um sowohl Bot als auch Web Interface zusammen zu starten, kannst du docker-compose erweitern:

```yaml
services:
  bot:
    # ... existing bot configuration
  
  web:
    build:
      context: ./web
      dockerfile: Dockerfile
    ports:
      - "5000:5000"
    volumes:
      - ./data:/app/data:ro
    environment:
      - FLASK_ENV=production
```

## Entwicklung

Der Bot verwendet die [discordgo](https://github.com/bwmarrin/discordgo) Library.

Füge neue Commands in `internal/commands/` hinzu und registriere sie in `internal/bot/bot.go`.

