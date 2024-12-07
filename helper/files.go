package helper

import (
	"io"
	"log"
	"mime/multipart"
	"os"
)

func UploadFile(file *multipart.File, path string, fileName string) error {
	defer (*file).Close()

	targetFile, err := os.OpenFile(path+"/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		return err
	}

	defer targetFile.Close()

	if _, err := io.Copy(targetFile, *file); err != nil {
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
	return file.Header.Get("Content-Type") == "image/jpeg" || file.Header.Get("Content-Type") == "image/png"
}

func CheckDirectory(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}
