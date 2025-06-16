package dataBase

import (
    "database/sql"
    _ "github.com/lib/pq" 
    "log"
)

func Conectar() *sql.DB {
    dsn := "postgres://postgres:246810@192.168.32.1:5432/finanzas?sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Error al conectar:", err)
    }

    return db
}
