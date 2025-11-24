# Deployment Guide - Prestige League Season Four

## Server Setup

### Voraussetzungen
- Ubuntu/Debian Server mit Root-Zugriff
- Docker & Docker Compose installiert
- Domain `prestigeleague.de` mit DNS auf Server IP zeigend

### 1. Docker Installation

```bash
# Docker installieren
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Docker Compose installieren
sudo apt-get update
sudo apt-get install docker-compose-plugin

# User zu Docker-Gruppe hinzufügen
sudo usermod -aG docker $USER
newgrp docker
```

### 2. Repository klonen

```bash
git clone <repository-url>
cd PrestigueLeagueSeasonFour
```

### 3. Umgebungsvariablen konfigurieren

```bash
# .env Datei erstellen
cp .env.example .env
nano .env
```

Fülle die `.env` Datei aus:
```env
DISCORD_BOT_TOKEN=your_actual_discord_bot_token
ACME_EMAIL=your-email@prestigeleague.de
```

### 4. Let's Encrypt Verzeichnis vorbereiten

```bash
# Verzeichnis erstellen mit korrekten Rechten
mkdir -p letsencrypt
chmod 600 letsencrypt
```

### 5. DNS Konfiguration

Stelle sicher, dass folgende DNS-Einträge auf deine Server-IP zeigen:
- `prestigeleague.de` → A Record → `<SERVER_IP>`
- `www.prestigeleague.de` → A Record → `<SERVER_IP>`
- `traefik.prestigeleague.de` → A Record → `<SERVER_IP>` (optional, für Dashboard)

### 6. Container starten

```bash
# Alle Container bauen und starten
docker-compose up -d

# Logs anschauen
docker-compose logs -f

# Nur bestimmte Services logs
docker-compose logs -f traefik
docker-compose logs -f web
docker-compose logs -f bot
```

### 7. Status prüfen

```bash
# Container Status
docker-compose ps

# SSL Zertifikat prüfen
curl -I https://prestigeleague.de

# Traefik Dashboard (falls aktiviert)
# https://traefik.prestigeleague.de
```

## Services

### Traefik (Reverse Proxy & SSL)
- **Port 80**: HTTP (automatisch zu HTTPS umgeleitet)
- **Port 443**: HTTPS mit automatischen Let's Encrypt Zertifikaten
- **Port 8080**: Traefik Dashboard (intern, über Domain erreichbar)

### Web (Flask Anwendung)
- **URL**: https://prestigeleague.de
- Zeigt Match-Ergebnisse und Tabellen aller Divisionen
- Automatisches SSL via Traefik

### Bot (Discord Bot)
- Läuft im Hintergrund
- Verbindung zur Discord API
- Verwendet shared SQLite Datenbank in `./data`

## Wartung

### Container neu starten
```bash
docker-compose restart
```

### Logs anzeigen
```bash
docker-compose logs -f [service-name]
```

### Container stoppen
```bash
docker-compose down
```

### Updates durchführen
```bash
git pull
docker-compose build
docker-compose up -d
```

### Datenbank Backup
```bash
# Backup erstellen
cp data/league.db data/league.db.backup-$(date +%Y%m%d)

# Alte Backups löschen (älter als 7 Tage)
find data/ -name "league.db.backup-*" -mtime +7 -delete
```

## Troubleshooting

### SSL Zertifikat wird nicht erstellt
1. DNS-Einträge prüfen: `dig prestigeleague.de`
2. Port 80 und 443 müssen offen sein
3. Traefik Logs prüfen: `docker-compose logs traefik`
4. Let's Encrypt Rate Limits prüfen

### Web Service startet nicht
```bash
docker-compose logs web
```

### Bot verbindet nicht
```bash
# Bot Token prüfen
docker-compose logs bot

# .env Datei prüfen
cat .env
```

### Traefik Dashboard aktivieren (mit Basic Auth)

```bash
# Password Hash erstellen
sudo apt-get install apache2-utils
htpasswd -nb admin yourpassword

# Output in docker-compose.yml unter traefik labels einfügen:
# - "traefik.http.routers.traefik.middlewares=auth"
# - "traefik.http.middlewares.auth.basicauth.users=admin:$$apr1$$..."
```

## Firewall Konfiguration

```bash
# UFW Firewall (Ubuntu)
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP
sudo ufw allow 443/tcp  # HTTPS
sudo ufw enable
```

## Monitoring

### Health Checks
```bash
# Web Service
curl -I https://prestigeleague.de

# Traefik
curl -I http://localhost:8080/ping
```

### Resource Usage
```bash
docker stats
```

## Backup Script

Erstelle ein Backup-Script `/root/backup-prestigeleague.sh`:

```bash
#!/bin/bash
BACKUP_DIR="/root/backups/prestigeleague"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Datenbank Backup
cp /path/to/PrestigueLeagueSeasonFour/data/league.db $BACKUP_DIR/league_$DATE.db

# Alte Backups löschen (älter als 30 Tage)
find $BACKUP_DIR -name "league_*.db" -mtime +30 -delete

echo "Backup completed: $DATE"
```

Mache es ausführbar und füge zu Cron hinzu:
```bash
chmod +x /root/backup-prestigeleague.sh
crontab -e
# Füge hinzu: 0 2 * * * /root/backup-prestigeleague.sh
```

## Support

Bei Problemen:
1. Logs prüfen: `docker-compose logs`
2. Container Status: `docker-compose ps`
3. DNS prüfen: `dig prestigeleague.de`
4. Ports prüfen: `netstat -tulpn | grep -E ':(80|443|8080)'`
