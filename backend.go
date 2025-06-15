package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/Kaoru-D/Test/control"
    "github.com/Kaoru-D/Test/dataBase"
    "os"
)

func main() {
    db := dataBase.Conectar()

    // Ejecutar archivo SQL solo una vez si quieres crear tablas/datos
    sqlBytes, err := os.ReadFile("sql/finanzas.sql")
    if err != nil {
        log.Fatal("No se pudo leer el SQL:", err)
    }
    db.Exec(string(sqlBytes))

    r := gin.Default()
    r.GET("/transacciones", control.ObtenerTransacciones(db))

    r.POST("/transacciones", control.GuardarTransaccion(db))
    r.Run(":5432") // http://localhost:8080/transacciones
}

   

