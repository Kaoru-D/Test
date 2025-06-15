package control

import (
    "database/sql"
    "net/http"
    "github.com/gin-gonic/gin"
    "time"
)

type Transaccion struct {
    AccountID   int     `json:"account_id"`
    Amount      float64 `json:"amount"`
    Currency    string  `json:"currency"`
    Type        string  `json:"type"` // ejemplo: "deposito", "retiro"
    Description string  `json:"description"`
}



func ObtenerTransacciones(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        rows, err := db.Query("SELECT * FROM transacciones")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()

        var trans []map[string]interface{}
        cols, _ := rows.Columns()
        for rows.Next() {
            colsData := make([]interface{}, len(cols))
            colsPtr := make([]interface{}, len(cols))
            for i := range colsData {
                colsPtr[i] = &colsData[i]
            }

            rows.Scan(colsPtr...)
            rowMap := make(map[string]interface{})
            for i, colName := range cols {
                rowMap[colName] = colsData[i]
            }
            trans = append(trans, rowMap)
        }

        c.JSON(http.StatusOK, trans)
    }
}

func GuardarTransaccion(c *gin.Context) {
    var t Transaccion

    if err := c.ShouldBindJSON(&t); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    db, err := ConectarBD()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al conectar con la base de datos"})
        return
    }
    defer db.Close()

    _, err = db.Exec(`
        INSERT INTO transactions (account_id, amount, currency, type, description, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, t.AccountID, t.Amount, t.Currency, t.Type, t.Description, time.Now())

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Transacción guardada exitosamente"})
}