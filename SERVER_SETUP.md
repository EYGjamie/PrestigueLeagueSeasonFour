# Server Deployment - Schritt für Schritt Anleitung

## 1. Server Vorbereitung

### SSH Verbindung zum Server
```bash
ssh root@your-server-ip
# oder
ssh username@your-server-ip
```

### Docker installieren (falls noch nicht vorhanden)
```bash
# Docker installieren
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Docker Compose Plugin installieren
sudo apt-get update
sudo apt-get install docker-compose-plugin

# Benutzer zur Docker-Gruppe hinzufügen (optional, wenn nicht root)
sudo usermod -aG docker $USER
newgrp docker

# Installation prüfen
docker --version
docker compose version
```

## 2. Dateien auf Server hochladen

### Option A: Mit Git (empfohlen)
```bash
# Auf dem Server
cd /opt  # oder /home/username
git clone https://github.com/your-username/PrestigueLeagueSeasonFour.git
cd PrestigueLeagueSeasonFour
```

### Option B: Mit SCP/SFTP
```bash
# Auf deinem lokalen PC (PowerShell)
# Gesamtes Verzeichnis hochladen
scp -r C:\Users\Jamie\Documents\GitHub\PrestigueLeagueSeasonFour root@your-server-ip:/opt/

# Oder mit WinSCP / FileZilla GUI
```

## 3. Konfiguration auf dem Server

### .env Datei erstellen
```bash
cd /opt/PrestigueLeagueSeasonFour  # oder dein Pfad

# .env Datei aus Template erstellen
cp .env.example .env

# Mit Nano editieren
nano .env
```

**Wichtig! Trage folgende Werte ein:**
```env
DISCORD_BOT_TOKEN=dein_echter_discord_bot_token_hier
ACME_EMAIL=deine-email@example.com
```

Speichern mit `CTRL+O`, `Enter`, dann `CTRL+X` zum Beenden.

### Verzeichnisse vorbereiten
```bash
# Let's Encrypt Verzeichnis erstellen
mkdir -p letsencrypt
chmod 600 letsencrypt

# Data Verzeichnis erstellen (falls nicht vorhanden)
mkdir -p data
ls
# Prüfen ob alle wichtigen Dateien da sind
ls -la
# Du solltest sehen: docker-compose.yml, Dockerfile, web/, data/, etc.
```

## 4. DNS Konfiguration

**WICHTIG: Bevor du die Container startest!**

Gehe zu deinem Domain-Anbieter (z.B. IONOS, Strato, etc.) und erstelle folgende DNS-Einträge:

| Typ | Name | Ziel | TTL |
|-----|------|------|-----|
| A   | prestigeleague.de | `DEINE_SERVER_IP` | 3600 |
| A   | www.prestigeleague.de | `DEINE_SERVER_IP` | 3600 |
| A   | traefik.prestigeleague.de | `DEINE_SERVER_IP` | 3600 |

**Warten bis DNS propagiert ist (5-30 Minuten):**
```bash
# Prüfen ob DNS funktioniert
dig prestigeleague.de
# oder
nslookup prestigeleague.de
```

## 5. Firewall konfigurieren

```bash
# UFW Firewall (Ubuntu/Debian)
sudo ufw allow 22/tcp    # SSH - WICHTIG, sonst sperrst du dich aus!
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable

# Status prüfen
sudo ufw status

# Oder mit iptables
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

## 6. Container starten

```bash
cd /opt/PrestigueLeagueSeasonFour

# Container bauen und starten
docker compose up -d

# Logs in Echtzeit anschauen
docker compose logs -f

# Oder nur einzelne Services
docker compose logs -f traefik   # SSL-Zertifikate
docker compose logs -f web       # Flask App
docker compose logs -f bot       # Discord Bot
```

## 7. Prüfen ob alles läuft

### Container Status
```bash
docker compose ps

# Du solltest 3 Container sehen:
# - prestigeleague-traefik (running)
# - prestigeleague-web (running)
# - prestigeleague-bot (running)
```

### SSL-Zertifikat prüfen
```bash
# Warte 1-2 Minuten nach Start
curl -I https://prestigeleague.de

# Du solltest sehen: HTTP/2 200
# Und: Server: Traefik
```

### Im Browser testen
Öffne: `https://prestigeleague.de`

Du solltest die Webseite mit grünem Schloss (SSL) sehen!

## 8. Discord Bot prüfen

```bash
# Bot Logs anschauen
docker compose logs bot

# Du solltest sehen:
# "Bot is now running"
# Keine Error Messages
```

Im Discord Server sollte der Bot jetzt online sein und Commands funktionieren:
- `/schedule division:1`
- `/createchannels matchday:1 category:123456789`
- `/report_result`

## Troubleshooting

### Problem: SSL-Zertifikat wird nicht erstellt

**Prüfen:**
```bash
# DNS korrekt?
dig prestigeleague.de

# Traefik Logs
docker compose logs traefik | grep -i error

# Ist Port 80/443 offen?
sudo netstat -tulpn | grep -E ':(80|443)'
```

**Lösung:**
- Warte 5-10 Minuten nach DNS-Änderung
- Prüfe ob Port 80/443 wirklich offen sind
- Prüfe ACME_EMAIL in .env Datei
- Let's Encrypt Rate Limits: Max 5 Versuche pro Stunde

### Problem: Bot startet nicht

```bash
docker compose logs bot

# Häufige Fehler:
# - "Invalid token" → DISCORD_BOT_TOKEN in .env prüfen
# - "Database error" → data/ Verzeichnis Rechte prüfen
```

**Lösung:**
```bash
# .env Datei prüfen
cat .env

# Container neu starten
docker compose restart bot

# Oder komplett neu bauen
docker compose down
docker compose up -d --build
```

### Problem: Webseite zeigt 502 Bad Gateway

```bash
# Web Container Logs
docker compose logs web

# Container Status
docker compose ps web
```

**Lösung:**
```bash
# Container neu starten
docker compose restart web

# Falls Fehler in Logs: Code-Fehler beheben und neu bauen
docker compose up -d --build web
```

### Problem: Container startet immer wieder neu

```bash
# Status mit Restart-Count
docker compose ps

# Letzte Logs vor Crash
docker compose logs --tail=50 web
```

**Lösung:**
- Logs lesen und Fehler beheben
- Oft: Fehlende Abhängigkeiten oder Code-Fehler

## 9. Wartung & Updates

### Logs anschauen
```bash
# Alle Logs
docker compose logs -f

# Letzte 100 Zeilen
docker compose logs --tail=100

# Nur Fehler
docker compose logs | grep -i error
```

### Container neu starten
```bash
# Alle Container
docker compose restart

# Nur einen Service
docker compose restart web
```

### Updates einspielen
```bash
cd /opt/PrestigueLeagueSeasonFour

# Git Pull (falls Git verwendet)
git pull

# Container neu bauen
docker compose down
docker compose up -d --build

# Oder ohne Downtime (Rolling Update)
docker compose up -d --build --no-deps web
```

### Datenbank Backup
```bash
# Backup erstellen
cp data/league.db data/league.db.backup-$(date +%Y%m%d)

# Automatisches Backup (siehe DEPLOYMENT.md)
```

### Volumes/Daten löschen (ACHTUNG!)
```bash
# Nur wenn du alles zurücksetzen willst
docker compose down -v  # Löscht auch Volumes!
rm -rf data/league.db letsencrypt/
```

## 10. Monitoring (optional)

### Ressourcen überwachen
```bash
# CPU, RAM, Netzwerk
docker stats

# Disk Space
df -h
```

### Automatische Restarts
Die Container haben bereits `restart: unless-stopped` konfiguriert.
Sie starten automatisch nach Server-Reboot!

### Service Status nach Reboot prüfen
```bash
# Nach Server-Neustart
docker compose ps

# Falls Container nicht laufen
docker compose up -d
```

## Zusammenfassung - Checkliste

- [ ] Docker installiert
- [ ] Dateien auf Server (Git Clone oder SCP)
- [ ] .env Datei erstellt und ausgefüllt
- [ ] DNS-Einträge gesetzt (A Records)
- [ ] Firewall Ports 80, 443 geöffnet
- [ ] `docker compose up -d` ausgeführt
- [ ] Logs geprüft: `docker compose logs -f`
- [ ] Webseite erreichbar: https://prestigeleague.de
- [ ] Bot online in Discord
- [ ] SSL-Zertifikat funktioniert (grünes Schloss)

## Nützliche Befehle - Spickzettel

```bash
# Container starten
docker compose up -d

# Container stoppen
docker compose down

# Logs live
docker compose logs -f

# Container neu starten
docker compose restart

# Status
docker compose ps

# Updates
git pull && docker compose up -d --build

# Backup
cp data/league.db data/league.db.backup
```

## Support

Bei Problemen:
1. Logs prüfen: `docker compose logs`
2. Container Status: `docker compose ps`
3. DNS prüfen: `dig prestigeleague.de`
4. Firewall prüfen: `sudo ufw status`

