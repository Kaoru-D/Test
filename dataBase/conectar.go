package dataBase

import (
    "database/sql"
    _ "github.com/lib/pq" // o el driver que uses
    "log"
)

func Conectar() *sql.DB {
    dsn := "user=postgres password=1234 dbname=finanzas sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Error al conectar:", err)
    }

    return db
}
