package control

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10" // Asegúrate de tener el paquete validator instalado
	"net/http"
	"time"
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

func ObtenerTransacciones(c *gin.Context) {
    db := c.MustGet("db").(*sql.DB)

    // Puedes permitir filtro por cuenta si quieres
    accountID := c.Query("account_id")

    query := `
        SELECT
            t.id,
            t.account_id,
            t.amount,
            t.currency,
            t.type,
            t.description,
            t.created_at,
            SUM(t.amount) OVER (
                PARTITION BY t.account_id
                ORDER BY t.created_at
                ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
            ) AS saldo
        FROM transactions t
        `

    var args []interface{}
    if accountID != "" {
        query += ` WHERE t.account_id = $1`
        args = append(args, accountID)
    }

    query += ` ORDER BY t.created_at DESC`

    rows, err := db.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var transacciones []Transaccion
    for rows.Next() {
        var t Transaccion
        if err := rows.Scan(&t.ID, &t.AccountID, &t.Amount, &t.Currency, &t.Type, &t.Description, &t.CreatedAt, &t.SaldoAcumulado); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        transacciones = append(transacciones, t)
    }

    c.JSON(http.StatusOK, transacciones)
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

func ListarTransacciones(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	// Puedes obtener el parámetro ?account_id=123 si es necesario
	accountID := c.Query("account_id")

	var rows *sql.Rows // Declaramos una variable para almacenar las filas obtenidas de la consulta
	var err error      // Declaramos una variable para manejar errores

	if accountID != "" { // Si se proporciona un account_id, filtramos las transacciones por cuenta
		// Ejecutamos una consulta SQL para obtener transacciones filtradas por account_id
		rows, err = db.Query(`
            SELECT id, account_id, amount, currency, type, description, created_at
            SUM(
                CASE 
                    WHEN type = 'deposito' THEN amount
                    WHEN type = 'retiro' THEN -amount
                    ELSE 0
                END
            ) OVER (
                PARTITION BY account_id 
                ORDER BY created_at
                ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
            ) AS saldo_acumulado
            FROM transactions
            WHERE account_id = $1
            ORDER BY created_at DESC
        `, accountID)
	} else { // Si no se proporciona account_id, obtenemos todas las transacciones
		// Ejecutamos una consulta SQL para obtener todas las transacciones
		startDate := c.Query("start_date") // Obtenemos el parámetro de fecha de inicio si se proporciona
		endDate := c.Query("end_date") // Obtenemos el parámetro de fecha de fin si se proporciona

		filtroFecha := "" // Inicializamos un filtro de fecha vacío
		params := []interface{}{}   // Creamos un slice para almacenar los parámetros de la consulta
		paramIndex := 1 // Inicializamos un índice para los parámetros

		if accountID != "" { // Si se proporciona un account_id, lo agregamos al filtro de fecha
			filtroFecha += fmt.Sprintf("account_id = $%d", paramIndex) // Agregamos el filtro de cuenta al filtro de fecha
			params = append(params, accountID) // Agregamos el account_id a los parámetros
			paramIndex++ // Incrementamos el índice del parámetro
		}

		if startDate != "" { // Si se proporciona una fecha de inicio, la agregamos al filtro de fecha
			if filtroFecha != "" { // Si ya hay un filtro de fecha, agregamos un AND
				filtroFecha += " AND " 
			}
			filtroFecha += fmt.Sprintf("created_at >= $%d", paramIndex) // Agregamos el filtro de fecha de inicio
			params = append(params, startDate) 
			paramIndex++
		}

		if endDate != "" { // Si se proporciona una fecha de fin, la agregamos al filtro de fecha
			if filtroFecha != "" { // Si ya hay un filtro de fecha, agregamos un AND
				filtroFecha += " AND "
			}
			filtroFecha += fmt.Sprintf("created_at <= $%d", paramIndex) // Agregamos el filtro de fecha de fin
			params = append(params, endDate) 
			paramIndex++
		}
        // Construimos la consulta SQL con el filtro de fecha si es necesario
		query := `
	    SELECT 
		id,
		account_id,
		amount,
		currency,
		type,
		description,
		created_at,
		SUM(
			CASE 
				WHEN type = 'deposito' THEN amount
				WHEN type = 'retiro' THEN -amount
				ELSE 0
			END
		) OVER (
			PARTITION BY account_id 
			ORDER BY created_at
			ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
		) AS saldo_acumulado
	FROM transactions
`

		if filtroFecha != "" {  // Si hay un filtro de fecha, lo agregamos a la consulta
			query += " WHERE " + filtroFecha 
		}

		query += " ORDER BY created_at DESC"    // Agregamos el ordenamiento por fecha de creación

		rows, err = db.Query(query, params...) // Ejecutamos la consulta con los parámetros correspondientes

	}

	// Si hay un error al ejecutar la consulta, respondemos con un error 500 Internal Server Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close() // Aseguramos que las filas se cierren al finalizar la función

	// Creamos un slice para almacenar las transacciones obtenidas
	var transacciones []Transaccion
	for rows.Next() { // Iteramos sobre las filas obtenidas
		var t Transaccion
		// Escaneamos los datos de la fila actual en la variable t

		// Asegúrate de que los tipos de datos coincidan con los de tu estructura Transaccion
		if err := rows.Scan(&t.ID, &t.AccountID, &t.Amount, &t.Currency, &t.Type, &t.Description, &t.CreatedAt, &t.SaldoAcumulado); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		transacciones = append(transacciones, t) // Agregamos la transacción al slice de transacciones
	}

	c.JSON(http.StatusOK, transacciones) // Respondemos con un JSON que contiene todas las transacciones obtenidas
}

