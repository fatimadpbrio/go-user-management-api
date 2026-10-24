package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	MaxFileSize = 5 << 20
	UploadPath  = "./uploads/profile"
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func UploadProfileImage(c *gin.Context) (string, error) {

    fileHeader, err := c.FormFile("profileImage")
    if err != nil {
        return "", nil
    }

    if fileHeader.Size > MaxFileSize {
        return "", errors.New("file size must be less than 5MB")
    }

    extension := strings.ToLower(filepath.Ext(fileHeader.Filename))

    if !allowedExtensions[extension] {
        return "", errors.New("invalid file format")
    }

    err = os.MkdirAll(UploadPath, 0755)
    if err != nil {
        return "", err
    }

    fileName := uuid.New().String() + extension

    filePath := filepath.Join(UploadPath, fileName)

    err = c.SaveUploadedFile(fileHeader, filePath)
    if err != nil {
        return "", err
    }

    return fileName, nil
}

func DeleteFile(filePath string) error {

	if filePath == "" {
		return errors.New("file name is empty")
	}


	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}