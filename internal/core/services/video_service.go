package services

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"
	"video-processor/internal/core/domain"
	"video-processor/internal/core/ports"
)

type videoJob struct {
	Video     *domain.Video
	VideoPath string
	Timestamp string
}

type videoService struct {
	processor ports.VideoProcessor
	storage   ports.Storage
	repo      ports.VideoRepository
	jobChan   chan videoJob
}

func NewVideoService(p ports.VideoProcessor, s ports.Storage, r ports.VideoRepository) ports.VideoUseCase {
	svc := &videoService{
		processor: p,
		storage:   s,
		repo:      r,
		jobChan:   make(chan videoJob, 100),
	}

	// Start worker pool (e.g., 3 workers)
	for i := 0; i < 3; i++ {
		go svc.worker()
	}

	return svc
}

func (s *videoService) worker() {
	for job := range s.jobChan {
		s.processJob(job)
	}
}

func (s *videoService) processJob(job videoJob) {
	video := job.Video
	video.Status = domain.StatusProcessing
	s.repo.Update(video)

	frames, err := s.processor.ExtractFrames(job.VideoPath, job.Timestamp)
	if err != nil {
		video.Status = domain.StatusFailed
		video.Message = "Erro no processamento: " + err.Error()
		s.repo.Update(video)
		s.storage.DeleteFile(job.VideoPath)
		return
	}

	zipFilename := fmt.Sprintf("frames_%s.zip", job.Timestamp)
	err = s.storage.SaveZip(zipFilename, frames)
	if err != nil {
		video.Status = domain.StatusFailed
		video.Message = "Erro ao criar ZIP: " + err.Error()
		s.repo.Update(video)
		s.storage.DeleteFile(job.VideoPath)
		return
	}

	// Cleanup
	s.storage.DeleteFile(job.VideoPath)
	if len(frames) > 0 {
		tempDir := filepath.Dir(frames[0])
		s.storage.DeleteDir(tempDir)
	}

	video.Status = domain.StatusCompleted
	video.ZipPath = zipFilename
	video.FrameCount = len(frames)
	video.Message = fmt.Sprintf("Processamento concluído! %d frames extraídos.", len(frames))
	s.repo.Update(video)

	log.Printf("Video %d processed successfully", video.ID)
}

func (s *videoService) UploadAndProcess(userID int64, filename string, file io.Reader) (domain.ProcessingResult, error) {
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

	video := &domain.Video{
		UserID:   userID,
		Filename: filename,
		Status:   domain.StatusPending,
	}

	err = s.repo.Create(video)
	if err != nil {
		s.storage.DeleteFile(videoPath)
		return domain.ProcessingResult{Success: false, Message: "Erro ao criar registro no banco: " + err.Error()}, err
	}

	// Queue for processing
	s.jobChan <- videoJob{
		Video:     video,
		VideoPath: videoPath,
		Timestamp: timestamp,
	}

	return domain.ProcessingResult{
		Success: true,
		Message: "Vídeo recebido e adicionado à fila de processamento!",
		VideoID: video.ID,
	}, nil
}

func (s *videoService) ListProcessedFiles() ([]domain.FileInfo, error) {
	return s.storage.ListOutputs()
}

func (s *videoService) GetVideosByUserID(userID int64) ([]domain.Video, error) {
	return s.repo.GetByUserID(userID)
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
