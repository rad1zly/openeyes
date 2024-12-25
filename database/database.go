package database

import (
    "database/sql"
    _"github.com/go-sql-driver/mysql"
)

func InitDB() (*sql.DB, error) {
    db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/database_name")
    if err != nil {
        return nil, err
    }
    return db, nil
}