package database

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "sync"

    "github.com/joho/godotenv"
    _ "github.com/go-sql-driver/mysql"
)

var (
    db   *sql.DB
    once sync.Once
)

func initDB() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")

    connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
    db, err = sql.Open("mysql", connectionString)
    if err != nil {
        log.Fatal(err)
    }

    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Connected to the database")
}

func GetDB() *sql.DB {
    once.Do(initDB)
    return db
}