// File: j-ticketing/pkg/utils/file_utils.go
package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// FileUtil handles file operations
type FileUtil struct {
	maxFileSize         int64
	allowedContentTypes []string
}

// NewFileUtil creates a new file utility with default settings
func NewFileUtil() *FileUtil {
	return &FileUtil{
		maxFileSize: 500 * 1024 * 1024, // 500MB
		allowedContentTypes: []string{
			"image/jpeg",
			"image/png",
		},
	}
}

// NewCustomFileUtil creates a new file utility with custom settings
func NewCustomFileUtil(maxFileSize int64, allowedContentTypes []string) *FileUtil {
	return &FileUtil{
		maxFileSize:         maxFileSize,
		allowedContentTypes: allowedContentTypes,
	}
}

// UploadAttachmentFile uploads a file to the specified storage path and returns the unique filename
func (f *FileUtil) UploadAttachmentFile(file *multipart.FileHeader, storagePath string) (string, error) {
	// Validate file
	if err := f.validateFile(file); err != nil {
		return "", err
	}

	// Sanitize the original filename
	sanitizedFilename := f.sanitizeFilename(file.Filename)

	// Generate unique filename
	uniqueFileName := uuid.New().String() + "-" + sanitizedFilename

	// Validate storage path
	if storagePath == "" {
		return "", errors.New("storage path cannot be empty")
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Save file
	targetPath := filepath.Join(storagePath, uniqueFileName)
	if err := f.saveUploadedFile(file, targetPath); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return uniqueFileName, nil
}

// sanitizeFilename removes or replaces problematic characters in filenames
func (f *FileUtil) sanitizeFilename(filename string) string {
	// Replace spaces with underscores
	sanitized := strings.ReplaceAll(filename, " ", "_")

	// Replace other problematic characters
	problematicChars := []string{
		"<", ">", ":", "\"", "|", "?", "*", "/", "\\",
		"\t", "\n", "\r", "\x00",
	}

	for _, char := range problematicChars {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}

	// Remove multiple consecutive underscores
	for strings.Contains(sanitized, "__") {
		sanitized = strings.ReplaceAll(sanitized, "__", "_")
	}

	// Trim underscores from beginning and end
	sanitized = strings.Trim(sanitized, "_")

	// Ensure filename is not empty after sanitization
	if sanitized == "" {
		sanitized = "file"
	}

	return sanitized
}

// DeleteAttachmentFile deletes a file by its unique filename from the specified storage path
func (f *FileUtil) DeleteAttachmentFile(uniqueFileName string, storagePath string) {
	if storagePath == "" || uniqueFileName == "" {
		return
	}

	filePath := filepath.Join(storagePath, uniqueFileName)
	os.Remove(filePath) // Ignore errors as file might not exist
}

// ValidateFile validates the uploaded file (public method for external validation)
func (f *FileUtil) ValidateFile(file *multipart.FileHeader) error {
	return f.validateFile(file)
}

// validateFile validates the uploaded file (private method)
func (f *FileUtil) validateFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > f.maxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %.0f MB", float64(f.maxFileSize)/(1024*1024))
	}

	// Check content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		return errors.New("content type not specified")
	}

	isAllowed := false
	for _, allowedType := range f.allowedContentTypes {
		if contentType == allowedType {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("file type not allowed. Allowed types: %v", f.allowedContentTypes)
	}

	return nil
}

// saveUploadedFile saves the uploaded file to the specified path
func (f *FileUtil) saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// GetMaxFileSize returns the maximum file size allowed
func (f *FileUtil) GetMaxFileSize() int64 {
	return f.maxFileSize
}

// GetAllowedContentTypes returns the list of allowed content types
func (f *FileUtil) GetAllowedContentTypes() []string {
	return f.allowedContentTypes
}

// SetMaxFileSize sets the maximum file size allowed
func (f *FileUtil) SetMaxFileSize(size int64) {
	f.maxFileSize = size
}

// SetAllowedContentTypes sets the allowed content types
func (f *FileUtil) SetAllowedContentTypes(types []string) {
	f.allowedContentTypes = types
}
