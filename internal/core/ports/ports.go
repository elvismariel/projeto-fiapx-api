package ports

import (
	"io"
	"video-processor/internal/core/domain"
)

// VideoUseCase is the Inbound Port
type VideoUseCase interface {
	UploadAndProcess(filename string, file io.Reader) (domain.ProcessingResult, error)
	ListProcessedFiles() ([]domain.FileInfo, error)
}

// VideoProcessor is the Outbound Port for video processing logic
type VideoProcessor interface {
	ExtractFrames(videoPath string, timestamp string) ([]string, error)
}

// Storage is the Outbound Port for file operations
type Storage interface {
	SaveUpload(filename string, data io.Reader) (string, error)
	SaveZip(zipFilename string, files []string) error
	DeleteFile(path string) error
	DeleteDir(path string) error
	ListOutputs() ([]domain.FileInfo, error)
	GetOutputPath(filename string) string
}

// UserUseCase is the Inbound Port for user logic
type UserUseCase interface {
	Register(email, password, name string) (domain.AuthResponse, error)
	Login(email, password string) (domain.AuthResponse, error)
}

// UserRepository is the Outbound Port for user data persistence
type UserRepository interface {
	Create(user *domain.User) error
	GetByEmail(email string) (*domain.User, error)
	GetByID(id int64) (*domain.User, error)
}
