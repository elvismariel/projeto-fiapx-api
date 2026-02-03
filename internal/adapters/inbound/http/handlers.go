package http

import (
	"net/http"
	"video-processor/internal/core/domain"
	"video-processor/internal/core/ports"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	videoUseCase ports.VideoUseCase
	userUseCase  ports.UserUseCase
	storage      ports.Storage
}

func NewHandler(v ports.VideoUseCase, u ports.UserUseCase, s ports.Storage) *Handler {
	return &Handler{
		videoUseCase: v,
		userUseCase:  u,
		storage:      s,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/", h.HandleIndex)
	r.POST("/upload", h.HandleVideoUpload)
	r.GET("/download/:filename", h.HandleDownload)
	r.GET("/api/status", h.HandleStatus)

	// Auth routes
	r.POST("/register", h.HandleRegister)
	r.POST("/login", h.HandleLogin)
}

func (h *Handler) HandleIndex(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, getHTMLForm())
}

func (h *Handler) HandleVideoUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Erro ao receber arquivo: " + err.Error()})
		return
	}
	defer file.Close()

	result, err := h.videoUseCase.UploadAndProcess(header.Filename, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) HandleDownload(c *gin.Context) {
	filename := c.Param("filename")
	filePath := h.storage.GetOutputPath(filename)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/zip")

	c.File(filePath)
}

func (h *Handler) HandleStatus(c *gin.Context) {
	files, err := h.videoUseCase.ListProcessedFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar arquivos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"total": len(files),
	})
}

func (h *Handler) HandleRegister(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Dados inválidos: " + err.Error()})
		return
	}

	response, err := h.userUseCase.Register(req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) HandleLogin(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Dados inválidos: " + err.Error()})
		return
	}

	response, err := h.userUseCase.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
