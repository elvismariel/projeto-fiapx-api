package services

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
	"video-processor/internal/core/domain"
	"video-processor/internal/core/ports"
)

type videoService struct {
	processor ports.VideoProcessor
	storage   ports.Storage
}

func NewVideoService(p ports.VideoProcessor, s ports.Storage) ports.VideoUseCase {
	return &videoService{
		processor: p,
		storage:   s,
	}
}

func (s *videoService) UploadAndProcess(filename string, file io.Reader) (domain.ProcessingResult, error) {
	if !s.isValidVideoFile(filename) {
		return domain.ProcessingResult{
			Success: false,
			Message: "Formato de arquivo não suportado. Use: mp4, avi, mov, mkv",
		}, nil
	}

	timestamp := time.Now().Format("20060102_150405")
	uniqueFilename := fmt.Sprintf("%s_%s", timestamp, filename)

	videoPath, err := s.storage.SaveUpload(uniqueFilename, file)
	if err != nil {
		return domain.ProcessingResult{Success: false, Message: "Erro ao salvar arquivo: " + err.Error()}, err
	}

	frames, err := s.processor.ExtractFrames(videoPath, timestamp)
	if err != nil {
		return domain.ProcessingResult{Success: false, Message: "Erro no processamento: " + err.Error()}, err
	}

	zipFilename := fmt.Sprintf("frames_%s.zip", timestamp)
	err = s.storage.SaveZip(zipFilename, frames)
	if err != nil {
		return domain.ProcessingResult{Success: false, Message: "Erro ao criar ZIP: " + err.Error()}, err
	}

	// Cleanup
	s.storage.DeleteFile(videoPath)
	if len(frames) > 0 {
		tempDir := filepath.Dir(frames[0])
		s.storage.DeleteDir(tempDir)
	}

	imageNames := make([]string, len(frames))
	for i, frame := range frames {
		imageNames[i] = filepath.Base(frame)
	}

	return domain.ProcessingResult{
		Success:    true,
		Message:    fmt.Sprintf("Processamento concluído! %d frames extraídos.", len(frames)),
		ZipPath:    zipFilename,
		FrameCount: len(frames),
		Images:     imageNames,
	}, nil
}

func (s *videoService) ListProcessedFiles() ([]domain.FileInfo, error) {
	return s.storage.ListOutputs()
}

func (s *videoService) isValidVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}
