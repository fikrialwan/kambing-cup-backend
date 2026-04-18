package helper

import (
	"bytes"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"github.com/adrium/goheif"
)

func UploadFile(reader io.Reader, path string, fileName string) error {
	targetFile, err := os.OpenFile(path+"/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return err
	}

	defer targetFile.Close()

	if _, err := io.Copy(targetFile, reader); err != nil {
		log.Printf("Error copying file: %v", err)
		return err
	}

	return nil
}

func DeleteFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("File does not exist: %v", err)
		return
	}

	if err := os.Remove(path); err != nil {
		log.Printf("Error deleting file: %v", err)
		return
	}
}

func IsImage(file *multipart.FileHeader) bool {
	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	return contentType == "image/jpeg" ||
		contentType == "image/jpg" ||
		contentType == "image/png" ||
		contentType == "image/heic" ||
		contentType == "image/heif"
}

func IsHEIC(file *multipart.FileHeader) bool {
	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	return contentType == "image/heic" || contentType == "image/heif"
}

func ConvertHEICToJPEG(file io.Reader) ([]byte, error) {
	img, err := goheif.Decode(file)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ValidateImageSize(file *multipart.FileHeader, maxSize int64) bool {
	return file.Size <= maxSize
}

func CheckDirectory(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}
