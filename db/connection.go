package db

import (
    "fmt"
    "log"
    "os"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

var DB *sqlx.DB

func Connect() *sqlx.DB {
    connStr := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
    )

    db, err := sqlx.Connect("postgres", connStr)
    if err != nil {
        log.Fatalf("❌ Database connection failed: %v", err)
    }
    log.Println("✅ Connected to PostgreSQL database:", os.Getenv("DB_NAME"))
    DB = db
    return db
}
