package main

import (
	"fmt"
	"log"
	inbound_http "video-processor/internal/adapters/inbound/http"
	outbound_processor "video-processor/internal/adapters/outbound/processor"
	outbound_storage "video-processor/internal/adapters/outbound/storage"
	core_services "video-processor/internal/core/services"

	"os/exec"

	"github.com/gin-gonic/gin"
)

func main() {
	// Verify if ffmpeg is installed
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Fatal("❌ Erro: ffmpeg não encontrado no sistema. Por favor, instale o ffmpeg para continuar.")
	}

	// Initialize Outbound Adapters
	storage := outbound_storage.NewFSStorage()
	processor := outbound_processor.NewFFmpegProcessor()

	// Initialize Core Service (Use Case)
	videoService := core_services.NewVideoService(processor, storage)

	// Initialize Inbound Adapter (HTTP)
	handler := inbound_http.NewHandler(videoService, storage)

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
