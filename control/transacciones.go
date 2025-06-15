package control

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
	"github.com/go-playground/validator/v10" // Asegúrate de tener el paquete validator instalado
	"github.com/gin-gonic/gin"
)

// Transaccion representa una transacción financiera
// con campos como ID de cuenta, monto, moneda, tipo y descripción
// Asegúrate de que los tipos de datos coincidan con los de tu base de datos
// Puedes agregar más campos según sea necesario
type Transaccion struct {
	/* AccountID   int     `json:"account_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Type        string  `json:"type"` // ejemplo: "deposito", "retiro"
	Description string  `json:"description"` */

	// Definimos los campos de la transacción con binding para validación
	AccountID   int     `json:"account_id" binding:"required,min=1"`           // Aseguramos que el ID de la cuenta sea requerido y mayor a 1
	Amount      float64 `json:"amount" binding:"required,gt=0"`                // Aseguramos que el monto sea requerido y mayor a 0
	Currency    string  `json:"currency" binding:"required,len=3"`             // Ej: USD, COP
	Type        string  `json:"type" binding:"required,oneof=deposito retiro"` // Aseguramos que el tipo sea requerido y solo pueda ser "deposito" o "retiro"
	Description string  `json:"description" binding:"required"`                // Aseguramos que la descripción sea requerida
}

func ObtenerTransacciones(db *sql.DB) gin.HandlerFunc {
	// Esta función maneja la obtención de transacciones desde la base de datos
	return func(c *gin.Context) {
		// Realizamos una consulta a la base de datos para obtener todas las transacciones
		rows, err := db.Query("SELECT * FROM transactions")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close() // Aseguramos que los recursos se liberen al finalizar la función

		// Creamos un slice para almacenar las transacciones obtenidas
		var trans []map[string]interface{}
		cols, _ := rows.Columns()
		for rows.Next() {
			colsData := make([]interface{}, len(cols))
			colsPtr := make([]interface{}, len(cols))
			for i := range colsData {
				colsPtr[i] = &colsData[i]
			}

			// Escaneamos los datos de la fila actual en colsData
			rows.Scan(colsPtr...)
			rowMap := make(map[string]interface{})
			for i, colName := range cols {
				rowMap[colName] = colsData[i]
			}
			trans = append(trans, rowMap) // Agregamos el mapa de la fila a nuestro slice de transacciones
		}

		c.JSON(http.StatusOK, trans) // Respondemos con un JSON que contiene todas las transacciones obtenidas
	}
}

func GuardarTransaccion(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Esta función maneja la creación de una nueva transacción
		// Se espera que el cuerpo de la solicitud contenga un JSON con los detalles de la transacción
		var t Transaccion // Definimos una variable de tipo Transaccion para almacenar los datos

		if err := c.ShouldBindJSON(&t); err != nil { // Intentamos vincular el JSON del cuerpo de la solicitud a la variable t
			// Si hay un error al vincular, respondemos con un error 400 Bad Request
			// y enviamos el mensaje de error al cliente
			var ve validator.ValidationErrors // Verificamos si el error es de validación

			if errors.As(err, &ve) { // Si el error es de validación, lo procesamos

				var errores []string // Creamos un slice para almacenar los mensajes de error

				// Mapa de nombres de campos → nombres amigables
				var nombresCampo = map[string]string{
					"AccountID":   "Cuenta",
					"Amount":      "Monto",
					"Currency":    "Moneda",
					"Type":        "Tipo de transacción",
					"Description": "Descripción",
				}

				for _, fe := range ve { // Iteramos sobre los errores de validación
                    campoOriginal := fe.Field() // Obtenemos el nombre del campo original

					campo := nombresCampo[campoOriginal] // Obtenemos el nombre amigable del campo, si no existe, usamos el original

					switch fe.Tag() {   // Verificamos el tipo de error

					case "required": // Si el campo es requerido y no se proporcionó
						errores = append(errores, fmt.Sprintf("El campo '%s' es requerido", campo))

					case "min": // Si el campo debe ser mayor a un valor mínimo
						errores = append(errores, fmt.Sprintf("El campo '%s' debe ser mayor o igual a %s", campo, fe.Param()))

					case "gt": // Si el campo debe ser mayor a 0
						errores = append(errores, fmt.Sprintf("El campo '%s' debe ser mayor que %s", campo, fe.Param()))

					case "len": // Si el campo debe tener una longitud específica
						errores = append(errores, fmt.Sprintf("El campo '%s' debe ser exactamente %s caracteres", campo , fe.Param()))

					case "oneof": // Si el campo debe ser uno de los valores especificados
						errores = append(errores, fmt.Sprintf("El campo '%s' debe ser uno de los siguientes: %s", campo , fe.Param()))

					default: // Si el error es de otro tipo, lo agregamos como un error genérico
						errores = append(errores, fmt.Sprintf("Error de validación en el campo '%s': %s", campo, fe.Tag()))
					}
				}
				// Si hay errores de validación, respondemos con un error 400 Bad Request
				c.JSON(http.StatusBadRequest, gin.H{"errores": errores}) // Respondemos con un JSON que contiene los errores de validación
				return
			}
			// Otro tipo de error que no es de validación
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Si la vinculación es exitosa, procedemos a insertar los datos en la base de datos
		// Ejecutamos una consulta SQL para insertar una nueva transacción
		_, err := db.Exec(`
            INSERT INTO transactions (account_id, amount, currency, type, description, created_at)
            VALUES ($1, $2, $3, $4, $5, $6)
        `, t.AccountID, t.Amount, t.Currency, t.Type, t.Description, time.Now())

		// Si hay un error al ejecutar la consulta, respondemos con un error 500 Internal Server Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Si la inserción es exitosa, respondemos con un mensaje de éxito 200 OK
		c.JSON(http.StatusOK, gin.H{"message": "Transacción guardada exitosamente"})
	}
}
