package main

import (
    "log"
    "os"
    "github.com/gin-gonic/gin"
    "github.com/Kaoru-D/Test/control"
    "github.com/Kaoru-D/Test/dataBase"
    
)

func main() {
    db := dataBase.Conectar()

    // Ejecutar archivo SQL solo una vez si quieres crear tablas/datos
    sqlBytes, err := os.ReadFile("sql/finanzas.sql")// Se asegura de que la ruta sea correcta
    if err != nil { // se asegura de que el archivo exista
        log.Fatal("No se pudo leer el SQL:", err)
    }
    db.Exec(string(sqlBytes))// Ejecuta el SQL para crear tablas o insertar datos

    r := gin.Default() // Crea un nuevo router
    
    // Middleware para inyectar la conexión a la base de datos en el contexto de Gin
    r.Use(func(c *gin.Context) {
        c.Set("db", db)
        c.Next()
    })

    r.GET("/transacciones", control.ListarTransacciones)   // Define la ruta para obtener transacciones
    r.POST("/transacciones", control.GuardarTransaccion(db)) // Define la ruta para guardar una transacción
    
    r.Run(":8080") // http://localhost:8080/transacciones 
}

   

