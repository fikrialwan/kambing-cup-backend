package helper_test

import (
	"kambing-cup-backend/helper"
	"mime/multipart"
	"net/textproto"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsImage(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"application/pdf", false},
		{"text/plain", false},
		{"", false},
	}

	for _, test := range tests {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Type", test.contentType)
		fileHeader := &multipart.FileHeader{
			Header: header,
		}
		result := helper.IsImage(fileHeader)
		assert.Equal(t, test.expected, result)
	}
}

func TestCheckDirectory(t *testing.T) {
	dir := "./test_dir"
	
	// Ensure cleanup
	defer os.Remove(dir)

	// Directory should not exist yet
	_, err := os.Stat(dir)
	assert.True(t, os.IsNotExist(err))

	helper.CheckDirectory(dir)

	// Directory should exist now
	info, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Call again should not panic/error
	helper.CheckDirectory(dir)
}

// UploadFile and DeleteFile involve file I/O which might be trickier to test cleanly without mocking filesystem or creating real temp files.
// For integration/unit testing with real files:

func TestFileOperations(t *testing.T) {
	dir := "./temp_test_files"
	helper.CheckDirectory(dir)
	defer os.RemoveAll(dir) // Cleanup

	// We skip UploadFile test here as creating a multipart.File from scratch is verbose,
	// but we can test DeleteFile.
	
	filePath := dir + "/test.txt"
	file, err := os.Create(filePath)
	assert.NoError(t, err)
	file.Close()

	// Verify exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	helper.DeleteFile(filePath)

	// Verify deleted
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))
	
	// Delete non-existent file should not panic
	helper.DeleteFile(filePath) 
}
