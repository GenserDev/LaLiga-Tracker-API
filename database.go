package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func initDatabase() *sql.DB {
	// Configuración de conexión simplificada
	connStr := "postgresql://usuario:contraseña@localhost/laliga?sslmode=disable"
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Crear tabla si no existe
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS matches (
			id SERIAL PRIMARY KEY,
			home_team TEXT NOT NULL,
			away_team TEXT NOT NULL,
			match_date DATE NOT NULL,
			goals INTEGER DEFAULT 0,
			yellow_cards INTEGER DEFAULT 0,
			red_cards INTEGER DEFAULT 0,
			extra_time BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conexión a base de datos establecida")
	return db
}