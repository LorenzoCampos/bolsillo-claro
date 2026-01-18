package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/LorenzoCampos/bolsillo-claro/internal/config"
	"github.com/LorenzoCampos/bolsillo-claro/internal/database"
	accountsHandler "github.com/LorenzoCampos/bolsillo-claro/internal/handlers/accounts"
	authHandler "github.com/LorenzoCampos/bolsillo-claro/internal/handlers/auth"
	categoriesHandler "github.com/LorenzoCampos/bolsillo-claro/internal/handlers/categories"
	dashboardHandler "github.com/LorenzoCampos/bolsillo-claro/internal/handlers/dashboard"
	expensesHandler "github.com/LorenzoCampos/bolsillo-claro/internal/handlers/expenses"
	incomesHandler "github.com/LorenzoCampos/bolsillo-claro/internal/handlers/incomes"
	savingsGoalsHandler "github.com/LorenzoCampos/bolsillo-claro/internal/handlers/savings_goals"
	"github.com/LorenzoCampos/bolsillo-claro/internal/middleware"
)

// Server encapsula el servidor HTTP y su configuración
type Server struct {
	config *config.Config // Puntero a la configuración
	router *gin.Engine    // El router de Gin que maneja las rutas HTTP
	db     *database.DB   // Puntero al pool de conexiones de PostgreSQL
}

// New crea una nueva instancia del servidor
// Recibe la configuración y la conexión a la DB, retorna un puntero a Server
func New(cfg *config.Config, db *database.DB) *Server {
	// Crear el router de Gin
	// gin.Default() incluye middleware de logging y recovery automático
	router := gin.Default()

	// Crear la instancia del servidor
	server := &Server{
		config: cfg,
		router: router,
		db:     db,
	}

	// Configurar las rutas
	server.setupRoutes()

	return server
}

// setupRoutes configura todas las rutas de la API
func (s *Server) setupRoutes() {
	// Crear handlers
	authH := authHandler.NewHandler(s.db, s.config)
	accountsH := accountsHandler.NewHandler(s.db)

	// Crear middlewares
	authMiddleware := middleware.AuthMiddleware(s.config.JWTSecret)
	accountMiddleware := middleware.AccountMiddleware(s.db)

	// Grupo de rutas para la API
	// Todas las rutas estarán bajo /api
	api := s.router.Group("/api")
	{
		// Ruta de health check para verificar que el servidor está vivo
		api.GET("/health", s.healthCheck)

		// Ruta de prueba
		api.GET("/hello", s.helloWorld)

		// Ruta con parámetro dinámico en la URL
		// :name captura el valor de la URL (ej: /api/greet/Juan captura "Juan")
		api.GET("/greet/:name", s.greetUser)

		// Ruta con query parameters
		// Ej: /api/calculate?a=5&b=3
		api.GET("/calculate", s.calculate)

		// Nueva ruta: verificar conexión a PostgreSQL
		api.GET("/db-test", s.testDatabase)

		// Rutas de autenticación (públicas - no requieren auth)
		// Aplicamos rate limiting agresivo para prevenir brute-force attacks
		authRoutes := api.Group("/auth")
		authRoutes.Use(middleware.RateLimitAuth()) // 5 intentos cada 15 minutos
		{
			authRoutes.POST("/register", authH.Register)
			authRoutes.POST("/login", authH.Login)
			authRoutes.POST("/refresh", authH.Refresh) // Renovar tokens con refresh token
		}

		// Rutas de cuentas (protegidas - requieren auth)
		accountsRoutes := api.Group("/accounts")
		accountsRoutes.Use(authMiddleware) // Aplicar middleware a todas las rutas del grupo
		{
			accountsRoutes.GET("/:id", accountsH.GetAccount)       // Obtener detalle de una cuenta
			accountsRoutes.PUT("/:id", accountsH.UpdateAccount)    // Actualizar cuenta
			accountsRoutes.DELETE("/:id", accountsH.DeleteAccount) // Eliminar cuenta
			accountsRoutes.GET("", accountsH.ListAccounts)         // Listar cuentas del usuario
			accountsRoutes.POST("", accountsH.CreateAccount)       // Crear nueva cuenta
		}

		// Rutas de gastos (protegidas - requieren auth + account)
		expensesRoutes := api.Group("/expenses")
		expensesRoutes.Use(authMiddleware)    // Primero validar autenticación
		expensesRoutes.Use(accountMiddleware) // Luego validar X-Account-ID
		{
			expensesRoutes.POST("", expensesHandler.CreateExpense(s.db.Pool))       // Crear gasto
			expensesRoutes.GET("/:id", expensesHandler.GetExpense(s.db.Pool))       // Obtener gasto por ID
			expensesRoutes.PUT("/:id", expensesHandler.UpdateExpense(s.db.Pool))    // Actualizar gasto
			expensesRoutes.DELETE("/:id", expensesHandler.DeleteExpense(s.db.Pool)) // Eliminar gasto
			expensesRoutes.GET("", expensesHandler.ListExpenses(s.db.Pool))         // Listar gastos
		}

		// Rutas de ingresos (protegidas - requieren auth + account)
		incomesRoutes := api.Group("/incomes")
		incomesRoutes.Use(authMiddleware)    // Primero validar autenticación
		incomesRoutes.Use(accountMiddleware) // Luego validar X-Account-ID
		{
			incomesRoutes.POST("", incomesHandler.CreateIncome(s.db.Pool))       // Crear ingreso
			incomesRoutes.GET("/:id", incomesHandler.GetIncome(s.db.Pool))       // Obtener ingreso por ID
			incomesRoutes.PUT("/:id", incomesHandler.UpdateIncome(s.db.Pool))    // Actualizar ingreso
			incomesRoutes.DELETE("/:id", incomesHandler.DeleteIncome(s.db.Pool)) // Eliminar ingreso
			incomesRoutes.GET("", incomesHandler.ListIncomes(s.db.Pool))         // Listar ingresos
		}

		// Rutas de categorías de gastos (protegidas - requieren auth + account)
		expenseCategoriesRoutes := api.Group("/expense-categories")
		expenseCategoriesRoutes.Use(authMiddleware)
		expenseCategoriesRoutes.Use(accountMiddleware)
		{
			expenseCategoriesRoutes.GET("", categoriesHandler.ListExpenseCategories(s.db.Pool))
			expenseCategoriesRoutes.POST("", categoriesHandler.CreateExpenseCategory(s.db.Pool))
			expenseCategoriesRoutes.PUT("/:id", categoriesHandler.UpdateExpenseCategory(s.db.Pool))
			expenseCategoriesRoutes.DELETE("/:id", categoriesHandler.DeleteExpenseCategory(s.db.Pool))
		}

		// Rutas de categorías de ingresos (protegidas - requieren auth + account)
		incomeCategoriesRoutes := api.Group("/income-categories")
		incomeCategoriesRoutes.Use(authMiddleware)
		incomeCategoriesRoutes.Use(accountMiddleware)
		{
			incomeCategoriesRoutes.GET("", categoriesHandler.ListIncomeCategories(s.db.Pool))
			incomeCategoriesRoutes.POST("", categoriesHandler.CreateIncomeCategory(s.db.Pool))
			incomeCategoriesRoutes.PUT("/:id", categoriesHandler.UpdateIncomeCategory(s.db.Pool))
			incomeCategoriesRoutes.DELETE("/:id", categoriesHandler.DeleteIncomeCategory(s.db.Pool))
		}

		// Rutas de dashboard (protegidas - requieren auth + account)
		dashboardRoutes := api.Group("/dashboard")
		dashboardRoutes.Use(authMiddleware)
		dashboardRoutes.Use(accountMiddleware)
		{
			dashboardRoutes.GET("/summary", dashboardHandler.GetSummary(s.db.Pool))
		}

		// Rutas de savings goals (protegidas - requieren auth + account)
		savingsGoalsRoutes := api.Group("/savings-goals")
		savingsGoalsRoutes.Use(authMiddleware)
		savingsGoalsRoutes.Use(accountMiddleware)
		{
			savingsGoalsRoutes.POST("", savingsGoalsHandler.CreateSavingsGoal(s.db.Pool))
			savingsGoalsRoutes.GET("", savingsGoalsHandler.ListSavingsGoals(s.db.Pool))
			savingsGoalsRoutes.GET("/:id", savingsGoalsHandler.GetSavingsGoal(s.db.Pool))
			savingsGoalsRoutes.PUT("/:id", savingsGoalsHandler.UpdateSavingsGoal(s.db.Pool))
			savingsGoalsRoutes.DELETE("/:id", savingsGoalsHandler.DeleteSavingsGoal(s.db.Pool))
			savingsGoalsRoutes.POST("/:id/add-funds", savingsGoalsHandler.AddFunds(s.db.Pool))
			savingsGoalsRoutes.POST("/:id/withdraw-funds", savingsGoalsHandler.WithdrawFunds(s.db.Pool))
		}
	}
}

// healthCheck es un endpoint simple que retorna el estado del servidor
// Los servicios de monitoreo usan este tipo de endpoints para verificar que la app funciona
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Bolsillo Claro API está funcionando correctamente",
	})
}

// helloWorld es un endpoint de prueba para verificar que todo funciona
func (s *Server) helloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "¡Hola desde Bolsillo Claro!",
		"info":    "Este es tu primer endpoint en Go con Gin",
	})
}

// greetUser demuestra cómo capturar parámetros de la URL
// :name en la ruta se captura con c.Param("name")
func (s *Server) greetUser(c *gin.Context) {
	// Capturar el parámetro :name de la URL
	name := c.Param("name")

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("¡Hola %s! Bienvenido a Bolsillo Claro", name),
		"name":    name,
	})
}

// calculate demuestra cómo capturar query parameters
// Query params son opcionales: /api/calculate?a=5&b=3
func (s *Server) calculate(c *gin.Context) {
	// c.Query() obtiene query parameters (retorna string vacío si no existe)
	// c.DefaultQuery() permite definir un valor por defecto
	a := c.DefaultQuery("a", "0")
	b := c.DefaultQuery("b", "0")

	// En Go, necesitamos convertir strings a números manualmente
	// Por ahora solo retornamos los valores recibidos
	c.JSON(http.StatusOK, gin.H{
		"operation": "suma",
		"a":         a,
		"b":         b,
		"message":   fmt.Sprintf("Recibí: a=%s y b=%s", a, b),
	})
}

// testDatabase hace una query simple a PostgreSQL para verificar conectividad
// Esto demuestra cómo ejecutar queries con pgx
func (s *Server) testDatabase(c *gin.Context) {
	// Crear un contexto para la query
	// c.Request.Context() usa el contexto de la HTTP request
	ctx := c.Request.Context()

	// Hacer ping a la base de datos
	err := s.db.Ping(ctx)
	if err != nil {
		// Si hay error, retornar 500 Internal Server Error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error conectando a PostgreSQL",
			"details": err.Error(),
		})
		return
	}

	// Ejecutar una query simple: SELECT NOW()::TEXT retorna la hora actual como texto
	var currentTime string
	query := "SELECT NOW()::TEXT"
	err = s.db.Pool.QueryRow(ctx, query).Scan(&currentTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error ejecutando query",
			"details": err.Error(),
		})
		return
	}

	// Obtener estadísticas del pool de conexiones
	stats := s.db.Stats()

	// Todo OK, retornar éxito con info de la DB
	c.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"message":       "Conexión a PostgreSQL exitosa",
		"database_time": currentTime,
		"pool_stats": gin.H{
			"total_connections":    stats.TotalConns(),
			"idle_connections":     stats.IdleConns(),
			"acquired_connections": stats.AcquiredConns(),
		},
	})
}

// Start inicia el servidor HTTP en el puerto configurado
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.config.Port)

	fmt.Printf("\n🚀 Servidor iniciado en http://localhost%s\n", addr)
	fmt.Printf("📍 Rutas disponibles:\n")
	fmt.Printf("   - GET  http://localhost%s/api/health\n", addr)
	fmt.Printf("   - GET  http://localhost%s/api/hello\n", addr)
	fmt.Printf("   - GET  http://localhost%s/api/greet/:name\n", addr)
	fmt.Printf("   - GET  http://localhost%s/api/calculate?a=5&b=3\n", addr)
	fmt.Printf("   - GET  http://localhost%s/api/db-test\n", addr)
	fmt.Printf("\n🔐 Autenticación:\n")
	fmt.Printf("   - POST http://localhost%s/api/auth/register\n", addr)
	fmt.Printf("   - POST http://localhost%s/api/auth/login\n", addr)
	fmt.Printf("   - POST http://localhost%s/api/auth/refresh (Renovar tokens)\n", addr)
	fmt.Printf("\n💰 Cuentas (requiere autenticación):\n")
	fmt.Printf("   - GET    http://localhost%s/api/accounts (Listar cuentas)\n", addr)
	fmt.Printf("   - GET    http://localhost%s/api/accounts/:id (Obtener detalle)\n", addr)
	fmt.Printf("   - POST   http://localhost%s/api/accounts (Crear cuenta)\n", addr)
	fmt.Printf("   - PUT    http://localhost%s/api/accounts/:id (Actualizar cuenta)\n", addr)
	fmt.Printf("   - DELETE http://localhost%s/api/accounts/:id (Eliminar cuenta)\n", addr)
	fmt.Printf("\n💸 Gastos (requiere autenticación + X-Account-ID):\n")
	fmt.Printf("   - GET    http://localhost%s/api/expenses (Listar gastos con filtros)\n", addr)
	fmt.Printf("   - GET    http://localhost%s/api/expenses/:id (Obtener detalle de gasto)\n", addr)
	fmt.Printf("   - POST   http://localhost%s/api/expenses (Registrar gasto)\n", addr)
	fmt.Printf("   - PUT    http://localhost%s/api/expenses/:id (Actualizar gasto)\n", addr)
	fmt.Printf("   - DELETE http://localhost%s/api/expenses/:id (Eliminar gasto)\n", addr)
	fmt.Printf("\n💰 Ingresos (requiere autenticación + X-Account-ID):\n")
	fmt.Printf("   - GET    http://localhost%s/api/incomes (Listar ingresos con filtros)\n", addr)
	fmt.Printf("   - GET    http://localhost%s/api/incomes/:id (Obtener detalle de ingreso)\n", addr)
	fmt.Printf("   - POST   http://localhost%s/api/incomes (Registrar ingreso)\n", addr)
	fmt.Printf("   - PUT    http://localhost%s/api/incomes/:id (Actualizar ingreso)\n", addr)
	fmt.Printf("   - DELETE http://localhost%s/api/incomes/:id (Eliminar ingreso)\n", addr)
	fmt.Printf("\n🏷️  Categorías (requiere autenticación + X-Account-ID):\n")
	fmt.Printf("   - GET    http://localhost%s/api/expense-categories (Listar categorías de gastos)\n", addr)
	fmt.Printf("   - POST   http://localhost%s/api/expense-categories (Crear categoría custom)\n", addr)
	fmt.Printf("   - PUT    http://localhost%s/api/expense-categories/:id (Actualizar)\n", addr)
	fmt.Printf("   - DELETE http://localhost%s/api/expense-categories/:id (Eliminar)\n", addr)
	fmt.Printf("   - GET    http://localhost%s/api/income-categories (Listar categorías de ingresos)\n", addr)
	fmt.Printf("   - POST   http://localhost%s/api/income-categories (Crear categoría custom)\n", addr)
	fmt.Printf("   - PUT    http://localhost%s/api/income-categories/:id (Actualizar)\n", addr)
	fmt.Printf("   - DELETE http://localhost%s/api/income-categories/:id (Eliminar)\n", addr)
	fmt.Printf("\n📊 Dashboard (requiere autenticación + X-Account-ID):\n")
	fmt.Printf("   - GET    http://localhost%s/api/dashboard/summary?month=YYYY-MM (Resumen financiero del mes)\n", addr)
	fmt.Printf("\n🎯 Metas de Ahorro (requiere autenticación + X-Account-ID):\n")
	fmt.Printf("   - GET    http://localhost%s/api/savings-goals (Listar metas)\n", addr)
	fmt.Printf("   - GET    http://localhost%s/api/savings-goals/:id (Detalle con historial)\n", addr)
	fmt.Printf("   - POST   http://localhost%s/api/savings-goals (Crear meta)\n", addr)
	fmt.Printf("   - PUT    http://localhost%s/api/savings-goals/:id (Actualizar meta)\n", addr)
	fmt.Printf("   - DELETE http://localhost%s/api/savings-goals/:id (Eliminar meta)\n", addr)
	fmt.Printf("   - POST   http://localhost%s/api/savings-goals/:id/add-funds (Agregar fondos)\n", addr)
	fmt.Printf("   - POST   http://localhost%s/api/savings-goals/:id/withdraw-funds (Retirar fondos)\n", addr)
	fmt.Println()

	// Iniciar el servidor
	// Run es bloqueante: el programa se queda acá escuchando peticiones
	return s.router.Run(addr)
}
