package archiver

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestUntar(t *testing.T) {
	dir, err := os.Getwd()
	fileDir, _ := filepath.Split(dir)
	log.Println(fileDir)

	testFilePath := filepath.Join(fileDir, "package.tar")
	err = Untar(testFilePath, dir)
	if err != nil {
		log.Println("Untar function failed!")
	}
}
