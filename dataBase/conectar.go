package dataBase

import (
    "database/sql"
    _ "github.com/lib/pq" 
    "log"
)

func Conectar() *sql.DB {
    dsn := "host=192.168.32.1 user=postgres password=246810 dbname=finanzas port=5432 sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Error al conectar:", err)
    }

    return db
}
