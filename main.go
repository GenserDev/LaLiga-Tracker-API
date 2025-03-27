package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var db *sql.DB

func main() {
	// Inicializar base de datos
	db = initDatabase()
	defer db.Close()

	// Definir rutas
	http.HandleFunc("/api/matches", handleMatches)
	http.HandleFunc("/api/matches/", handleMatchByID)

	// Iniciar servidor
	fmt.Println("Servidor iniciado en :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleMatches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch r.Method {
	case "GET":
		matches, err := GetAllMatches(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(matches)

	case "POST":
		var match Match
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &match)
		match.MatchDate = time.Now() // Usar fecha actual si no se especifica
		
		err := match.Create(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(match)

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func handleMatchByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Extraer ID de la URL
	id, _ := strconv.Atoi(r.URL.Path[len("/api/matches/"):])

	switch r.Method {
	case "GET":
		match, err := GetMatchByID(db, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(match)

	case "PUT":
		var match Match
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &match)
		match.ID = id
		
		err := match.Update(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(match)

	case "DELETE":
		err := DeleteMatch(db, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}