package dataBase

import (
    "database/sql"
    _ "github.com/lib/pq" 
    "log"
)

func Conectar() *sql.DB {
    dsn := "user=postgres password=246810 dbname=finanzas sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Error al conectar:", err)
    }

    return db
}
