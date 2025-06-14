package control

import (
    "database/sql"
    "net/http"
    "github.com/gin-gonic/gin"
)

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
