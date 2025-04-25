package config

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "82eab2ea-bbbc-4025-9858-5fa2738ec7e4.hsvc.ir"
	port     = 32215
	user     = "postgres"
	password = "KwuujJsBtnS4cTbowdpZYjdlKy8Vk0Du"
	dbname   = "postgres"
)

func InitDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return db, nil
}
