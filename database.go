package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func initDatabase() *sql.DB {
	// Obtener variables de entorno con valores por defecto
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "usuario"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "contraseña"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "laliga"
	}

	// Crear cadena de conexión
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	
	// Abrir conexión
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error al conectar a la base de datos:", err)
	}

	// Verificar conexión
	err = db.Ping()
	if err != nil {
		log.Fatal("Error al hacer ping a la base de datos:", err)
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
		log.Fatal("Error al crear tabla:", err)
	}

	fmt.Println("Conexión a base de datos establecida exitosamente")
	return db
}