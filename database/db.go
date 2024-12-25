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

func InitDB() (*sql.DB, error) {
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
    db, err := sql.Open("mysql", connectionString)
    if err != nil {
        return nil, err
    }

    err = db.Ping()
    if err != nil {
        return nil, err
    }

    log.Println("Connected to the database")
    return db, nil
}

func GetDB() *sql.DB {
    once.Do(func() {
        var err error
        db, err = InitDB()
        if err != nil {
            log.Fatal("Failed to initialize database:", err)
        }
    })
    return db
}