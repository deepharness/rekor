package provider

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

type DBProvider struct {
	DB *sql.DB
}

// NewDBProvider initializes the database connection
func NewDBProvider(dsn string) *DBProvider {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	// Ping the database to ensure connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging the database: %v", err)
	}

	fmt.Println("Connected to the database successfully.")
	return &DBProvider{DB: db}
}

// Close closes the database connection
func (p *DBProvider) Close() {
	if p.DB != nil {
		err := p.DB.Close()
		if err != nil {
			log.Printf("Error closing the database: %v", err)
		}
		fmt.Println("Database connection closed.")
	}
}
