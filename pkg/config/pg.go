package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type EndpointPostgres struct {
	Host     string
	Password string
	Username string
	Port     int
	Database string
}

func ConnectPostgres(endpoint EndpointPostgres) (*sql.DB, error) {
	// Membuat string koneksi
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		endpoint.Host, endpoint.Port, endpoint.Username, endpoint.Password, endpoint.Database)

	// Membuka koneksi ke database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database: ", err)
		return nil, err
	}

	// Memeriksa koneksi
	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting to database: ", err)
		return nil, err
	}

	return db, nil
}
