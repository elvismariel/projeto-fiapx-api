package main

import (
	"fmt"
	"log"
	inbound_http "video-processor/internal/adapters/inbound/http"
	outbound_processor "video-processor/internal/adapters/outbound/processor"
	outbound_repository "video-processor/internal/adapters/outbound/repository"
	outbound_storage "video-processor/internal/adapters/outbound/storage"
	core_services "video-processor/internal/core/services"

	"context"
	"os"

	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	fmt.Println("Starting Video Processor Server v2 (Worker Pool)...")
	// Verify if ffmpeg is installed
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Fatal("❌ Erro: ffmpeg não encontrado no sistema. Por favor, instale o ffmpeg para continuar.")
	}

	// Database initialization
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		dbHost = "localhost"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	// Run migrations
	m, err := migrate.New("file://migrations", connStr)
	if err != nil {
		log.Printf("⚠️ Erro ao preparar migrações: %v", err)
	} else {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Printf("⚠️ Erro ao aplicar migrações: %v", err)
		} else {
			fmt.Println("✅ Migrações aplicadas com sucesso!")
		}
	}

	dbPool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal("❌ Erro ao conectar ao banco de dados: ", err)
	}
	defer dbPool.Close()

	// Initialize Outbound Adapters
	storage := outbound_storage.NewFSStorage()
	processor := outbound_processor.NewFFmpegProcessor()
	userRepo := outbound_repository.NewPostgresUserRepository(dbPool)
	videoRepo := outbound_repository.NewPostgresVideoRepository(dbPool)

	// Initialize Core Services
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "fiapx-secret-key"
	}

	videoService := core_services.NewVideoService(processor, storage, videoRepo)
	userService := core_services.NewUserService(userRepo, jwtSecret)

	// Initialize Inbound Adapter (HTTP)
	handler := inbound_http.NewHandler(videoService, userService, storage, jwtSecret)

	r := gin.Default()

	// Middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Static files
	r.Static("/uploads", "./uploads")
	r.Static("/outputs", "./outputs")

	// Register Routes
	handler.RegisterRoutes(r)

	fmt.Println("🎬 Servidor iniciado na porta 8080")
	fmt.Println("📂 Acesse: http://localhost:8080")

	log.Fatal(r.Run(":8080"))
}
