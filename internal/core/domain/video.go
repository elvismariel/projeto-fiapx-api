package domain

type ProcessingResult struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	ZipPath    string   `json:"zip_path,omitempty"`
	FrameCount int      `json:"frame_count,omitempty"`
	Images     []string `json:"images,omitempty"`
}

type FileInfo struct {
	Name        string `json:"filename"`
	Size        int64  `json:"size"`
	CreatedAt   string `json:"created_at"`
	DownloadURL string `json:"download_url"`
}
