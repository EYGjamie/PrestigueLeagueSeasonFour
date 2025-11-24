import sqlite3
from flask import Flask, render_template, jsonify, send_from_directory
from pathlib import Path
import os

app = Flask(__name__)

# Pfad zur Datenbank
# Im Container ist die DB unter /app/data gemountet
if os.path.exists('/app/data/league.db'):
    DB_PATH = Path('/app/data/league.db')
else:
    # Lokale Entwicklung
    DB_PATH = Path(__file__).parent.parent / 'data' / 'league.db'

DATA_DIR = Path(__file__).parent / 'bg'


def get_db():
    """Erstellt eine Datenbankverbindung"""
    # Prüfe ob Datenbank existiert
    if not DB_PATH.exists():
        raise FileNotFoundError(f"Datenbank nicht gefunden: {DB_PATH}. Bitte erst den Bot mit --import starten oder migrate ausführen.")
    
    conn = sqlite3.connect(str(DB_PATH))
    conn.row_factory = sqlite3.Row
    return conn


def calculate_standings(division):
    """Berechnet die Tabelle für eine Division"""
    conn = get_db()
    cursor = conn.cursor()
    
    # Alle Teams der Division
    cursor.execute("""
        SELECT id, name FROM teams WHERE division = ? ORDER BY name
    """, (division,))
    teams = {row['id']: {
        'name': row['name'],
        'played': 0,
        'wins': 0,
        'losses': 0,
        'games_won': 0,
        'games_lost': 0,
        'points': 0
    } for row in cursor.fetchall()}
    
    # Alle abgeschlossenen Matches
    cursor.execute("""
        SELECT team_home_id, team_away_id, score_home, score_away
        FROM matches
        WHERE division = ? AND score_home IS NOT NULL AND score_away IS NOT NULL
    """, (division,))
    
    for match in cursor.fetchall():
        home_id = match['team_home_id']
        away_id = match['team_away_id']
        score_home = match['score_home']
        score_away = match['score_away']
        
        if away_id is None or away_id == 0:  # Free Win
            continue
            
        # Statistiken aktualisieren
        teams[home_id]['played'] += 1
        teams[away_id]['played'] += 1
        teams[home_id]['games_won'] += score_home
        teams[home_id]['games_lost'] += score_away
        teams[away_id]['games_won'] += score_away
        teams[away_id]['games_lost'] += score_home
        
        if score_home > score_away:
            teams[home_id]['wins'] += 1
            teams[home_id]['points'] += 3
            teams[away_id]['losses'] += 1
        else:
            teams[away_id]['wins'] += 1
            teams[away_id]['points'] += 3
            teams[home_id]['losses'] += 1
    
    conn.close()
    
    # Sortieren nach Punkten, dann Spielen gewonnen
    standings = sorted(teams.values(), 
                      key=lambda x: (x['points'], x['games_won'] - x['games_lost'], x['games_won']), 
                      reverse=True)
    
    return standings


def get_matches_by_division(division):
    """Holt alle Matches einer Division"""
    conn = get_db()
    cursor = conn.cursor()
    
    cursor.execute("""
        SELECT 
            m.id, m.matchday, m.score_home, m.score_away, m.reported_at,
            ht.name as home_team, at.name as away_team
        FROM matches m
        JOIN teams ht ON m.team_home_id = ht.id
        LEFT JOIN teams at ON m.team_away_id = at.id
        WHERE m.division = ?
        ORDER BY m.matchday, m.id
    """, (division,))
    
    matches = []
    for row in cursor.fetchall():
        # Überspringe Spiele, bei denen ein Team NULL, "0" oder nicht vorhanden ist
        if not row['away_team'] or row['away_team'] == '0':
            continue
        if not row['home_team'] or row['home_team'] == '0':
            continue
            
        matches.append({
            'id': row['id'],
            'matchday': row['matchday'],
            'home_team': row['home_team'],
            'away_team': row['away_team'],
            'score_home': row['score_home'],
            'score_away': row['score_away'],
            'reported_at': row['reported_at'],
            'completed': row['score_home'] is not None
        })
    
    conn.close()
    return matches


@app.route('/')
def index():
    """Startseite mit Übersicht aller Divisionen"""
    conn = get_db()
    cursor = conn.cursor()
    
    cursor.execute("SELECT DISTINCT division FROM teams ORDER BY division")
    divisions = [row['division'] for row in cursor.fetchall()]
    
    conn.close()
    
    return render_template('index.html', divisions=divisions)


@app.route('/division/<int:division>')
def division_view(division):
    """Zeigt Tabelle und Matches einer Division"""
    standings = calculate_standings(division)
    matches = get_matches_by_division(division)
    
    # Gruppiere Matches nach Spieltag
    matchdays = {}
    for match in matches:
        md = match['matchday']
        if md not in matchdays:
            matchdays[md] = []
        matchdays[md].append(match)
    
    return render_template('division.html', 
                          division=division, 
                          standings=standings,
                          matchdays=sorted(matchdays.items()))


@app.route('/api/standings/<int:division>')
def api_standings(division):
    """API Endpoint für Standings"""
    standings = calculate_standings(division)
    return jsonify(standings)


@app.route('/api/matches/<int:division>')
def api_matches(division):
    """API Endpoint für Matches"""
    matches = get_matches_by_division(division)
    return jsonify(matches)


if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=5000)
