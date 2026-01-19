package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/logger"
)

// RecurringIncomeTemplate representa un template activo que puede generar ingresos
type RecurringIncomeTemplate struct {
	ID                        string
	AccountID                 string
	Description               string
	Amount                    float64
	Currency                  string
	CategoryID                *string
	FamilyMemberID            *string
	RecurrenceFrequency       string
	RecurrenceInterval        int
	RecurrenceDayOfMonth      *int
	RecurrenceDayOfWeek       *int
	StartDate                 time.Time
	EndDate                   *time.Time
	TotalOccurrences          *int
	CurrentOccurrence         int
	ExchangeRate              *float64
	AmountInPrimaryCurrency   *float64
}

// GenerateDailyRecurringIncomes genera ingresos recurrentes para el día de hoy
// Debe ejecutarse UNA VEZ por día (idealmente a las 00:00)
func GenerateDailyRecurringIncomes(pool *pgxpool.Pool) error {
	ctx := context.Background()
	today := time.Now().UTC().Truncate(24 * time.Hour)

	logger.Info("scheduler.recurring_incomes.start", "Iniciando generación diaria de ingresos recurrentes", map[string]interface{}{
		"date": today.Format("2006-01-02"),
	})

	// Obtener templates activos que necesitan generar ingresos HOY
	templates, err := getIncomeTemplatesForToday(pool, ctx, today)
	if err != nil {
		logger.Error("scheduler.recurring_incomes.error", "Error obteniendo templates", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	if len(templates) == 0 {
		logger.Info("scheduler.recurring_incomes.complete", "No hay templates para generar hoy", map[string]interface{}{
			"date": today.Format("2006-01-02"),
		})
		return nil
	}

	logger.Info("scheduler.recurring_incomes.found", fmt.Sprintf("Encontrados %d templates para procesar", len(templates)), map[string]interface{}{
		"count": len(templates),
	})

	// Procesar cada template
	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, template := range templates {
		// Verificar si ya se generó un ingreso para este template HOY
		alreadyGenerated, err := checkIfIncomeAlreadyGenerated(pool, ctx, template.ID, today)
		if err != nil {
			logger.Error("scheduler.recurring_incomes.check_error", "Error verificando duplicados", map[string]interface{}{
				"template_id": template.ID,
				"error":       err.Error(),
			})
			errorCount++
			continue
		}

		if alreadyGenerated {
			logger.Info("scheduler.recurring_incomes.skip", "Gasto ya generado hoy (skip)", map[string]interface{}{
				"template_id": template.ID,
				"description": template.Description,
				"date":        today.Format("2006-01-02"),
			})
			skipCount++
			continue
		}

		// Calcular la fecha del ingreso a generar
		incomeDate := calculateIncomeGenerationDate(template, today)

		// Generar el ingreso
		err = generateActualIncomeFromTemplate(pool, ctx, template, incomeDate)
		if err != nil {
			logger.Error("scheduler.recurring_incomes.generate_error", "Error generando ingreso", map[string]interface{}{
				"template_id": template.ID,
				"description": template.Description,
				"error":       err.Error(),
			})
			errorCount++
			continue
		}

		// Incrementar current_occurrence
		err = incrementIncomeOccurrence(pool, ctx, template)
		if err != nil {
			logger.Error("scheduler.recurring_incomes.increment_error", "Error incrementando occurrence", map[string]interface{}{
				"template_id": template.ID,
				"error":       err.Error(),
			})
			// No marcamos como error porque el ingreso SÍ se generó
		}

		// Verificar si debemos desactivar el template (llegó al límite)
		shouldDeactivate := false
		deactivateReason := ""

		// Razón 1: Llegó a total_occurrences
		if template.TotalOccurrences != nil && template.CurrentOccurrence+1 >= *template.TotalOccurrences {
			shouldDeactivate = true
			deactivateReason = "total_occurrences reached"
		}

		// Razón 2: Llegó a end_date
		if template.EndDate != nil && !incomeDate.Before(*template.EndDate) {
			shouldDeactivate = true
			if deactivateReason != "" {
				deactivateReason += " + end_date reached"
			} else {
				deactivateReason = "end_date reached"
			}
		}

		if shouldDeactivate {
			err = deactivateIncomeTemplate(pool, ctx, template.ID, deactivateReason)
			if err != nil {
				logger.Error("scheduler.recurring_incomes.deactivate_error", "Error desactivando template", map[string]interface{}{
					"template_id": template.ID,
					"reason":      deactivateReason,
					"error":       err.Error(),
				})
			} else {
				logger.Info("scheduler.recurring_incomes.deactivated", "Template desactivado automáticamente", map[string]interface{}{
					"template_id": template.ID,
					"description": template.Description,
					"reason":      deactivateReason,
				})
			}
		}

		successCount++
	}

	logger.Info("scheduler.recurring_incomes.complete", "Generación diaria completada", map[string]interface{}{
		"total":   len(templates),
		"success": successCount,
		"skipped": skipCount,
		"errors":  errorCount,
	})

	return nil
}

// getIncomeTemplatesForToday obtiene templates activos que deben generar un ingreso HOY
func getIncomeTemplatesForToday(pool *pgxpool.Pool, ctx context.Context, today time.Time) ([]RecurringIncomeTemplate, error) {
	query := `
		SELECT 
			id, account_id, description, amount, currency,
			category_id, family_member_id,
			recurrence_frequency, recurrence_interval,
			recurrence_day_of_month, recurrence_day_of_week,
			start_date, end_date,
			total_occurrences, current_occurrence,
			exchange_rate, amount_in_primary_currency
		FROM recurring_incomes
		WHERE is_active = true
		  AND start_date <= $1
		  AND (end_date IS NULL OR end_date >= $1)
		  AND (total_occurrences IS NULL OR current_occurrence < total_occurrences)
	`

	rows, err := pool.Query(ctx, query, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []RecurringIncomeTemplate

	for rows.Next() {
		var t RecurringIncomeTemplate
		var startDate, endDate interface{}

		err := rows.Scan(
			&t.ID, &t.AccountID, &t.Description, &t.Amount, &t.Currency,
			&t.CategoryID, &t.FamilyMemberID,
			&t.RecurrenceFrequency, &t.RecurrenceInterval,
			&t.RecurrenceDayOfMonth, &t.RecurrenceDayOfWeek,
			&startDate, &endDate,
			&t.TotalOccurrences, &t.CurrentOccurrence,
			&t.ExchangeRate, &t.AmountInPrimaryCurrency,
		)
		if err != nil {
			return nil, err
		}

		// Parsear dates
		if startDate != nil {
			t.StartDate, _ = time.Parse("2006-01-02", fmt.Sprint(startDate))
		}
		if endDate != nil {
			parsed, _ := time.Parse("2006-01-02", fmt.Sprint(endDate))
			t.EndDate = &parsed
		}

		// Filtrar por lógica de frecuencia (solo si debe generar HOY)
		if shouldGenerateIncomeToday(t, today) {
			templates = append(templates, t)
		}
	}

	return templates, nil
}

// shouldGenerateIncomeToday determina si un template debe generar un ingreso HOY
func shouldGenerateIncomeToday(t RecurringIncomeTemplate, today time.Time) bool {
	switch t.RecurrenceFrequency {
	case "daily":
		// Daily: genera todos los días (respetando interval)
		daysSinceStart := int(today.Sub(t.StartDate).Hours() / 24)
		return daysSinceStart%t.RecurrenceInterval == 0

	case "weekly":
		// Weekly: solo si hoy es el día de semana configurado
		if t.RecurrenceDayOfWeek == nil {
			return false
		}
		weekday := int(today.Weekday()) // 0=Sunday, 6=Saturday
		if weekday != *t.RecurrenceDayOfWeek {
			return false
		}
		// Verificar interval (cada N semanas)
		weeksSinceStart := int(today.Sub(t.StartDate).Hours() / (24 * 7))
		return weeksSinceStart%t.RecurrenceInterval == 0

	case "monthly":
		// Monthly: solo si hoy es el día del mes configurado
		if t.RecurrenceDayOfMonth == nil {
			return false
		}
		
		// Edge case: día 31 en meses cortos → último día del mes
		targetDay := *t.RecurrenceDayOfMonth
		lastDayOfMonth := time.Date(today.Year(), today.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if targetDay > lastDayOfMonth {
			targetDay = lastDayOfMonth
		}

		if today.Day() != targetDay {
			return false
		}

		// Verificar interval (cada N meses)
		monthsSinceStart := (today.Year()-t.StartDate.Year())*12 + int(today.Month()-t.StartDate.Month())
		return monthsSinceStart%t.RecurrenceInterval == 0

	case "yearly":
		// Yearly: solo si hoy es el mismo día/mes que start_date
		if t.RecurrenceDayOfMonth == nil {
			return false
		}
		if today.Month() != t.StartDate.Month() {
			return false
		}

		// Edge case: 29 de febrero en años no bisiestos → 28 de febrero
		targetDay := *t.RecurrenceDayOfMonth
		lastDayOfMonth := time.Date(today.Year(), today.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		if targetDay > lastDayOfMonth {
			targetDay = lastDayOfMonth
		}

		if today.Day() != targetDay {
			return false
		}

		// Verificar interval (cada N años)
		yearsSinceStart := today.Year() - t.StartDate.Year()
		return yearsSinceStart%t.RecurrenceInterval == 0

	default:
		return false
	}
}

// checkIfIncomeAlreadyGenerated verifica si ya se generó un ingreso para este template en esta fecha
func checkIfIncomeAlreadyGenerated(pool *pgxpool.Pool, ctx context.Context, templateID string, date time.Time) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM incomes
			WHERE recurring_income_id = $1
			  AND date = $2
		)
	`
	var exists bool
	err := pool.QueryRow(ctx, query, templateID, date).Scan(&exists)
	return exists, err
}

// calculateIncomeGenerationDate calcula la fecha del ingreso a generar
func calculateIncomeGenerationDate(t RecurringIncomeTemplate, today time.Time) time.Time {
	// Para la mayoría de casos, la fecha es HOY
	// Excepción: si el template empezó en el pasado y estamos haciendo catchup
	// Por ahora: siempre generamos con fecha de HOY
	return today
}

// generateActualIncomeFromTemplate crea un income desde un template
func generateActualIncomeFromTemplate(pool *pgxpool.Pool, ctx context.Context, t RecurringIncomeTemplate, incomeDate time.Time) error {
	insertQuery := `
		INSERT INTO incomes (
			account_id, family_member_id, category_id,
			description, amount, currency,
			exchange_rate, amount_in_primary_currency,
			income_type, date,
			recurring_income_id,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id
	`

	// Exchange rate: usar del template o default 1.0
	exchangeRate := 1.0
	if t.ExchangeRate != nil {
		exchangeRate = *t.ExchangeRate
	}

	// Amount in primary currency: usar del template o calcular
	amountInPrimaryCurrency := t.Amount
	if t.AmountInPrimaryCurrency != nil {
		amountInPrimaryCurrency = *t.AmountInPrimaryCurrency
	}

	var incomeID string
	err := pool.QueryRow(
		ctx,
		insertQuery,
		t.AccountID,
		t.FamilyMemberID,
		t.CategoryID,
		t.Description,
		t.Amount,
		t.Currency,
		exchangeRate,
		amountInPrimaryCurrency,
		"recurring", // income_type
		incomeDate,
		t.ID, // recurring_income_id (FK al template)
	).Scan(&incomeID)

	if err != nil {
		return err
	}

	logger.Info("scheduler.income.generated", "Gasto generado desde template", map[string]interface{}{
		"income_id":           incomeID,
		"recurring_income_id": t.ID,
		"account_id":           t.AccountID,
		"description":          t.Description,
		"amount":               t.Amount,
		"currency":             t.Currency,
		"date":                 incomeDate.Format("2006-01-02"),
	})

	return nil
}

// incrementIncomeOccurrence incrementa el contador current_occurrence del template
func incrementIncomeOccurrence(pool *pgxpool.Pool, ctx context.Context, t RecurringIncomeTemplate) error {
	query := "UPDATE recurring_incomes SET current_occurrence = current_occurrence + 1 WHERE id = $1"
	_, err := pool.Exec(ctx, query, t.ID)
	return err
}

// deactivateIncomeTemplate desactiva un template (soft delete)
func deactivateIncomeTemplate(pool *pgxpool.Pool, ctx context.Context, templateID string, reason string) error {
	query := "UPDATE recurring_incomes SET is_active = false WHERE id = $1"
	_, err := pool.Exec(ctx, query, templateID)
	return err
}
