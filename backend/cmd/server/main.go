package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/LorenzoCampos/bolsillo-claro/internal/config"
	"github.com/LorenzoCampos/bolsillo-claro/internal/database"
	"github.com/LorenzoCampos/bolsillo-claro/internal/server"
	"github.com/LorenzoCampos/bolsillo-claro/pkg/scheduler"
)

// main es la función especial que Go ejecuta al iniciar el programa
// Es el punto de entrada de toda aplicación Go
func main() {
	fmt.Println("🏦 Iniciando Bolsillo Claro API...")

	// Paso 1: Cargar la configuración desde variables de entorno
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Error cargando configuración: %v", err)
	}
	fmt.Println("✅ Configuración cargada correctamente")

	// Paso 2: Conectar a PostgreSQL
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Error conectando a PostgreSQL: %v", err)
	}
	// defer ejecuta la función al FINAL de main(), antes de que el programa termine
	// Esto garantiza que siempre cerremos el pool de conexiones
	defer db.Close()

	// Paso 3: Crear el servidor HTTP (ahora le pasamos también la DB)
	srv := server.New(cfg, db)
	fmt.Println("✅ Servidor HTTP creado")

	// Paso 3.5: Iniciar CRON scheduler para gastos e ingresos recurrentes
	c := cron.New()
	
	// Ejecutar generación diaria a las 00:01 (1 minuto después de medianoche)
	// Formato: "1 0 * * *" = minuto 1, hora 0, todos los días
	c.AddFunc("1 0 * * *", func() {
		fmt.Println("🔁 Ejecutando generación diaria de gastos recurrentes...")
		err := scheduler.GenerateDailyRecurringExpenses(db.Pool)
		if err != nil {
			log.Printf("❌ Error en generación de gastos recurrentes: %v", err)
		}
		
		fmt.Println("💰 Ejecutando generación diaria de ingresos recurrentes...")
		err = scheduler.GenerateDailyRecurringIncomes(db.Pool)
		if err != nil {
			log.Printf("❌ Error en generación de ingresos recurrentes: %v", err)
		}
	})
	
	// Iniciar CRON
	c.Start()
	fmt.Println("✅ CRON scheduler iniciado (ejecuta diariamente a las 00:01 UTC)")
	
	// Ejecutar una vez al arrancar el servidor (catchup de hoy si es necesario)
	go func() {
		fmt.Println("🔁 Ejecutando generación inicial de gastos (catchup)...")
		err := scheduler.GenerateDailyRecurringExpenses(db.Pool)
		if err != nil {
			log.Printf("❌ Error en generación inicial de gastos: %v", err)
		}
		
		fmt.Println("💰 Ejecutando generación inicial de ingresos (catchup)...")
		err = scheduler.GenerateDailyRecurringIncomes(db.Pool)
		if err != nil {
			log.Printf("❌ Error en generación inicial de ingresos: %v", err)
		}
	}()

	// Paso 4: Setup de graceful shutdown
	// Esto permite que el servidor se apague limpiamente cuando recibe SIGINT (Ctrl+C) o SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Paso 5: Iniciar el servidor en una goroutine (hilo ligero)
	// para que no bloquee y podamos escuchar señales de shutdown
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("❌ Error iniciando el servidor: %v", err)
		}
	}()

	// Esperar señal de shutdown
	<-quit
	fmt.Println("\n🛑 Señal de shutdown recibida, cerrando servidor...")
	
	// Detener CRON scheduler
	c.Stop()
	fmt.Println("✅ CRON scheduler detenido")
}
