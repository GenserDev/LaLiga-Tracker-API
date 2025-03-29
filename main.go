package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

// Match representa un partido de La Liga
type Match struct {
	ID           int       `json:"id"`
	HomeTeam     string    `json:"homeTeam"`
	AwayTeam     string    `json:"awayTeam"`
	MatchDate    string    `json:"matchDate"`
	HomeGoals    int       `json:"homeGoals"`
	AwayGoals    int       `json:"awayGoals"`
	YellowCards  int       `json:"yellowCards"`
	RedCards     int       `json:"redCards"`
	ExtraTime    int       `json:"extraTime"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

var db *sql.DB

func main() {
	// Inicializar la base de datos
	initDB()

	// Crear el router
	router := mux.NewRouter()

	// Definir API endpoints
	router.HandleFunc("/api/matches", getMatches).Methods("GET")
	router.HandleFunc("/api/matches/{id}", getMatch).Methods("GET")
	router.HandleFunc("/api/matches", createMatch).Methods("POST")
	router.HandleFunc("/api/matches/{id}", updateMatch).Methods("PUT")
	router.HandleFunc("/api/matches/{id}", deleteMatch).Methods("DELETE")

	router.HandleFunc("/api/matches/{id}/goals", updateGoals).Methods("PATCH")
	router.HandleFunc("/api/matches/{id}/yellowcards", updateYellowCards).Methods("PATCH")
	router.HandleFunc("/api/matches/{id}/redcards", updateRedCards).Methods("PATCH")
	router.HandleFunc("/api/matches/{id}/extratime", updateExtraTime).Methods("PATCH")

	// Configurar CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Usar middleware CORS
	handler := c.Handler(router)

	// Determinar puerto a usar
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Puerto por defecto
	}

	// Iniciar el servidor
	fmt.Printf("Servidor iniciado en http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// Inicializar la base de datos SQLite
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./laliga.db")
	if err != nil {
		log.Fatalf("Error abriendo base de datos: %v", err)
	}

	// Crear tabla de partidos 
	statement := `
	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		home_team TEXT NOT NULL,
		away_team TEXT NOT NULL,
		match_date TEXT NOT NULL,
		home_goals INTEGER DEFAULT 0,
		away_goals INTEGER DEFAULT 0,
		yellow_cards INTEGER DEFAULT 0,
		red_cards INTEGER DEFAULT 0,
		extra_time INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(statement)
	if err != nil {
		log.Fatalf("Error creando tabla: %v", err)
	}
}

// GET Obtener todos los partidos
func getMatches(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, home_team, away_team, match_date, 
						  home_goals, away_goals, yellow_cards, red_cards, extra_time FROM matches`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var m Match
		err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate, 
						&m.HomeGoals, &m.AwayGoals, &m.YellowCards, &m.RedCards, &m.ExtraTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		matches = append(matches, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

// GET Obtener un partido por ID
func getMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var m Match
	err := db.QueryRow(`SELECT id, home_team, away_team, match_date, 
					   home_goals, away_goals, yellow_cards, red_cards, extra_time FROM matches WHERE id = ?`, id).
		Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate, 
			&m.HomeGoals, &m.AwayGoals, &m.YellowCards, &m.RedCards, &m.ExtraTime)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Partido no encontrado", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// POST Crear un nuevo partido
func createMatch(w http.ResponseWriter, r *http.Request) {
	var m Match
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if m.HomeTeam == "" || m.AwayTeam == "" || m.MatchDate == "" {
		http.Error(w, "Equipo local, visitante y fecha son obligatorios", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`INSERT INTO matches (home_team, away_team, match_date) VALUES (?, ?, ?)`,
		m.HomeTeam, m.AwayTeam, m.MatchDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

// PUT Actualizar un partido existente
func updateMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var m Match
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if m.HomeTeam == "" || m.AwayTeam == "" || m.MatchDate == "" {
		http.Error(w, "Equipo local, visitante y fecha son obligatorios", http.StatusBadRequest)
		return
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM matches WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}

	_, err = db.Exec(`UPDATE matches SET home_team = ?, away_team = ?, match_date = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		m.HomeTeam, m.AwayTeam, m.MatchDate, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	idInt, _ := strconv.Atoi(id)
	m.ID = idInt
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// DELETE Eliminar un partido
func deleteMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM matches WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}

	_, err = db.Exec("DELETE FROM matches WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


// PATCH Actualizar goles de un partido
func updateGoals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	type GoalsData struct {
		HomeGoals int `json:"homeGoals"`
		AwayGoals int `json:"awayGoals"`
	}

	var data GoalsData
	if r.ContentLength == 0 {
		data.HomeGoals = 1
		data.AwayGoals = 0
	} else {
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	_, err := db.Exec(`UPDATE matches SET home_goals = home_goals + ?, away_goals = away_goals + ?, 
					  updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		data.HomeGoals, data.AwayGoals, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var m Match
	err = db.QueryRow(`SELECT id, home_team, away_team, match_date, 
					  home_goals, away_goals, yellow_cards, red_cards, extra_time FROM matches WHERE id = ?`, id).
		Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate, 
			&m.HomeGoals, &m.AwayGoals, &m.YellowCards, &m.RedCards, &m.ExtraTime)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// PATCH Registrar tarjeta amarilla
func updateYellowCards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := db.Exec(`UPDATE matches SET yellow_cards = yellow_cards + 1, 
					  updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var m Match
	err = db.QueryRow(`SELECT id, home_team, away_team, match_date, 
					  home_goals, away_goals, yellow_cards, red_cards, extra_time FROM matches WHERE id = ?`, id).
		Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate, 
			&m.HomeGoals, &m.AwayGoals, &m.YellowCards, &m.RedCards, &m.ExtraTime)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// PATCH Registrar tarjeta roja
func updateRedCards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := db.Exec(`UPDATE matches SET red_cards = red_cards + 1, 
					  updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var m Match
	err = db.QueryRow(`SELECT id, home_team, away_team, match_date, 
					  home_goals, away_goals, yellow_cards, red_cards, extra_time FROM matches WHERE id = ?`, id).
		Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate, 
			&m.HomeGoals, &m.AwayGoals, &m.YellowCards, &m.RedCards, &m.ExtraTime)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// PATCH Establecer tiempo extra
func updateExtraTime(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	type ExtraTimeData struct {
		Minutes int `json:"minutes"`
	}

	var data ExtraTimeData
	if r.ContentLength == 0 {
		data.Minutes = 5
	} else {
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	_, err := db.Exec(`UPDATE matches SET extra_time = ?, 
					  updated_at = CURRENT_TIMESTAMP WHERE id = ?`, data.Minutes, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var m Match
	err = db.QueryRow(`SELECT id, home_team, away_team, match_date, 
					  home_goals, away_goals, yellow_cards, red_cards, extra_time FROM matches WHERE id = ?`, id).
		Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate, 
			&m.HomeGoals, &m.AwayGoals, &m.YellowCards, &m.RedCards, &m.ExtraTime)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}