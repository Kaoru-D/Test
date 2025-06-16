package control

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10" // Asegúrate de tener el paquete validator instalado
)

/*
Transaccion representa una transacción financiera
con campos como ID de cuenta, monto, moneda, tipo y descripción
Asegúrate de que los tipos de datos coincidan con los de tu base de datos
Puedes agregar más campos según sea necesario
*/
type Transaccion struct {
	/* AccountID   int     `json:"account_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Type        string  `json:"type"` // ejemplo: "deposito", "retiro"
	Description string  `json:"description"` */

	// Definimos los campos de la transacción con binding para validación
	ID             int       `json:"id"`                                            // Aseguramos que el ID sea requerido y mayor a 0
	AccountID      int       `json:"account_id" binding:"required,min=1"`           // Aseguramos que el ID de la cuenta sea requerido y mayor a 1
	Amount         float64   `json:"amount" binding:"required,gt=0"`                // Aseguramos que el monto sea requerido y mayor a 0
	Currency       string    `json:"currency" binding:"required,len=3"`             // Ej: USD, COP
	Type           string    `json:"type" binding:"required,oneof=deposito retiro"` // Aseguramos que el tipo sea requerido y solo pueda ser "deposito" o "retiro"
	Description    string    `json:"description" binding:"required"`                // Aseguramos que la descripción sea requerida
	CreatedAt      time.Time `json:"created_at"`                                    // Fecha de creación de la transacción
	SaldoAcumulado float64   `json:"saldo_acumulado"`                               // Saldo acumulado después de la transacción
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

					switch fe.Tag() { // Verificamos el tipo de error

					case "required": // Si el campo es requerido y no se proporcionó
						errores = append(errores, fmt.Sprintf("El campo '%s' es requerido", campo))

					case "min": // Si el campo debe ser mayor a un valor mínimo
						errores = append(errores, fmt.Sprintf("El campo '%s' debe ser mayor o igual a %s", campo, fe.Param()))

					case "gt": // Si el campo debe ser mayor a 0
						errores = append(errores, fmt.Sprintf("El campo '%s' debe ser mayor que %s", campo, fe.Param()))

					case "len": // Si el campo debe tener una longitud específica
						errores = append(errores, fmt.Sprintf("El campo '%s' debe ser exactamente %s caracteres", campo, fe.Param()))

					case "oneof": // Si el campo debe ser uno de los valores especificados
						errores = append(errores, fmt.Sprintf("El campo '%s' debe ser uno de los siguientes: %s", campo, fe.Param()))

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
            INSERT INTO transactions (account_id, amount::float8, currency, type, description, created_at)
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

func ListarTransacciones(c *gin.Context) {// Esta función maneja la obtención de transacciones
    // Se espera que la solicitud contenga parámetros de consulta opcionales
	db := c.MustGet("db").(*sql.DB) // Obtenemos la conexión a la base de datos desde el contexto de Gin

	accountID := c.Query("account_id")  // Obtenemos el ID de la cuenta desde los parámetros de consulta
	startDate := c.Query("start_date")  // Obtenemos la fecha de inicio desde los parámetros de consulta
	endDate := c.Query("end_date")  // Obtenemos la fecha de fin desde los parámetros de consulta

	// Parámetros y condiciones
	var conditions []string    // Creamos un slice para almacenar las condiciones de la consulta
	var params []interface{} // Creamos un slice para almacenar los parámetros de la consulta
	paramIndex := 1 // Usamos un índice para los parámetros de consulta

	if accountID != "" { // Si se proporciona un ID de cuenta, lo agregamos a las condiciones
		conditions = append(conditions, fmt.Sprintf("t.account_id = $%d", paramIndex)) 
		params = append(params, accountID)
		paramIndex++
	}
	if startDate != "" { // Si se proporciona una fecha de inicio, la agregamos a las condiciones
		conditions = append(conditions, fmt.Sprintf("t.created_at >= $%d", paramIndex))
		params = append(params, startDate)
		paramIndex++
	}
	if endDate != "" { // Si se proporciona una fecha de fin, la agregamos a las condiciones
		conditions = append(conditions, fmt.Sprintf("t.created_at <= $%d", paramIndex))
		params = append(params, endDate)
		paramIndex++
	}

	// Base de la consulta
	query := `
		SELECT
			t.id,
			t.account_id,
			t.amount::float8,
			t.currency,
			t.type,
			t.description,
			t.created_at,
			SUM(
				CASE 
					WHEN t.type = 'deposito' THEN t.amount::float8
					WHEN t.type = 'retiro' THEN -t.amount::float8
					ELSE 0
				END
			) OVER (
				PARTITION BY t.account_id 
				ORDER BY t.created_at
				ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
			) AS saldo_acumulado
		FROM transactions t
	`

	// Condiciones WHERE
	if len(conditions) > 0 {
		query += " WHERE " + fmt.Sprint(conditions[0])
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	query += " ORDER BY t.created_at DESC" // Ordenamos por fecha de creación en orden descendente

	// Ejecutar consulta
	rows, err := db.Query(query, params...)
	if err != nil {
		log.Println("Error al obtener transacciones:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener transacciones"})
		return
	}
	defer rows.Close() // Aseguramos cerrar las filas después de usarlas

	var transacciones []Transaccion // Creamos un slice para almacenar las transacciones obtenidas
	for rows.Next() { // Iteramos sobre las filas obtenidas
		var t Transaccion // Creamos una variable de tipo Transaccion para almacenar los datos de cada fila
        // Escaneamos los datos de la fila en la variable t
		if err := rows.Scan(&t.ID, &t.AccountID, &t.Amount, &t.Currency, &t.Type, &t.Description, &t.CreatedAt, &t.SaldoAcumulado); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		transacciones = append(transacciones, t) // Agregamos la transacción al slice de transacciones
	}

	c.JSON(http.StatusOK, transacciones) // Respondemos con un JSON que contiene las transacciones obtenidas
}
