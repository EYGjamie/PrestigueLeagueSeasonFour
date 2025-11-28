# Datenbank Verwaltung

Die Prestige League verwendet SQLite als Datenbank. Es gibt mehrere Möglichkeiten, die Datenbank manuell zu verwalten:

## 1. Adminer (Web-Interface) ⭐ Empfohlen

Adminer ist ein webbasiertes Datenbank-Management-Tool, das bereits in `docker-compose.yml` integriert ist.

### Zugriff:
- **URL:** https://db.prestigeleague.de
- **Server:** SQLite 3
- **Database:** `/data/prestigeleague.db` (Pfad in Adminer)
- **Username:** (leer lassen für SQLite)
- **Password:** (leer lassen für SQLite)

### Authentifizierung einrichten:
Erstelle einen Hash für Basic Auth:
```bash
# Passwort Hash generieren (z.B. für Benutzer "admin" mit Passwort "deinpasswort")
htpasswd -nb admin deinpasswort
```

Füge das Ergebnis zu deiner `.env` Datei hinzu:
```env
ADMINER_AUTH=admin:$$apr1$$...
```

### Container starten:
```bash
docker-compose up -d adminer
```

## 2. SQLite CLI im Docker Container

Direkter Zugriff über die Kommandozeile im laufenden Bot-Container:

```bash
# In den Bot-Container wechseln
docker exec -it prestigeleague-bot sh

# SQLite CLI starten
sqlite3 /app/data/prestigeleague.db

# Beispiel-Befehle:
.tables                          # Alle Tabellen anzeigen
.schema teams                    # Schema einer Tabelle anzeigen
SELECT * FROM teams;             # Alle Teams anzeigen
SELECT * FROM matches;           # Alle Matches anzeigen
.quit                            # SQLite verlassen
```

## 3. Lokaler SQLite Browser

Du kannst die Datenbank auch direkt mit einem lokalen Tool öffnen:

1. **DB Browser for SQLite** (kostenlos)
   - Download: https://sqlitebrowser.org/
   - Öffne die Datei: `./data/prestigeleague.db`

2. **DBeaver** (kostenlos, Universal)
   - Download: https://dbeaver.io/
   - Verbinde mit SQLite und wähle die Datenbankdatei

## 4. Direkte Dateisystem-Zugriff

Die Datenbank-Datei liegt im `./data` Verzeichnis:
```
./data/prestigeleague.db
```

Du kannst diese Datei:
- Sichern (kopieren)
- Mit einem SQLite-Tool öffnen
- Ersetzen (z.B. nach einem Restore)

## Nützliche SQL Queries

### Teams verwalten
```sql
-- Alle Teams anzeigen
SELECT * FROM teams ORDER BY division, name;

-- Team hinzufügen
INSERT INTO teams (name, division, role_id) 
VALUES ('Neues Team', 1, 'discord_role_id');

-- Team aktualisieren
UPDATE teams SET division = 2 WHERE name = 'Team Name';

-- Team löschen
DELETE FROM teams WHERE name = 'Team Name';
```

### Matches verwalten
```sql
-- Alle Matches eines Spieltags anzeigen
SELECT 
    m.id,
    m.matchday,
    th.name as home_team,
    ta.name as away_team,
    m.score_home,
    m.score_away
FROM matches m
JOIN teams th ON m.team_home_id = th.id
LEFT JOIN teams ta ON m.team_away_id = ta.id
WHERE m.division = 1 AND m.matchday = 1
ORDER BY m.id;

-- Match-Ergebnis eintragen
UPDATE matches 
SET score_home = 3, score_away = 1, 
    reported_at = CURRENT_TIMESTAMP,
    reported_by = 'Admin'
WHERE id = 1;

-- Match zurücksetzen
UPDATE matches 
SET score_home = NULL, score_away = NULL,
    reported_at = NULL, reported_by = NULL
WHERE id = 1;
```

### Statistiken
```sql
-- Tabelle einer Division berechnen
SELECT 
    t.name,
    COUNT(CASE WHEN m.score_home > m.score_away THEN 1 END) as wins,
    COUNT(CASE WHEN m.score_home = m.score_away THEN 1 END) as draws,
    COUNT(CASE WHEN m.score_home < m.score_away THEN 1 END) as losses,
    SUM(COALESCE(m.score_home, 0)) as goals_for,
    SUM(COALESCE(m.score_away, 0)) as goals_against
FROM teams t
LEFT JOIN matches m ON t.id = m.team_home_id AND m.score_home IS NOT NULL
WHERE t.division = 1
GROUP BY t.id, t.name
ORDER BY wins DESC, goals_for DESC;
```

## Backup & Restore

### Backup erstellen
```bash
# Lokales Backup
cp ./data/prestigeleague.db ./data/backup_$(date +%Y%m%d_%H%M%S).db

# Oder im Container
docker exec prestigeleague-bot cp /app/data/prestigeleague.db /app/data/backup.db
```

### Restore
```bash
# Container stoppen
docker-compose stop bot web

# Datenbank ersetzen
cp ./data/backup.db ./data/prestigeleague.db

# Container neu starten
docker-compose up -d bot web
```

## Sicherheitshinweise

⚠️ **Wichtig:**
- Aktiviere **immer** Basic Auth für Adminer in Production
- Erstelle regelmäßige Backups vor größeren Änderungen
- Teste SQL-Queries erst in einer Backup-Kopie
- Achte auf die Foreign Key Constraints (Teams <-> Matches)
