package common

import (
	"errors"
	"fmt"
	"github.com/ricochet2200/go-disk-usage/du"
	"log"
	"math"
	"os"
	"path/filepath"
)

const GB = 1000 * 1000 * 1000

func ByteToGB(bytes int64) float64 {
	gb := float64(bytes) / GB
	return math.Trunc(gb*100) / 100
}

func closeFileChecked(file *os.File) {
	err := file.Close()
	if err != nil {
		fmt.Printf("Error closing file: %s\n", err)
	}
}

func GetAvailableDiskSpace(localPath string) uint64 {
	usage := du.NewDiskUsage(localPath)
	if usage == nil {
		log.Fatalf("Could not get disk usage for path: %s", localPath)
	}
	return usage.Available() // bytes
}

func VerifyTarget(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Target download path %s does not exist.\n", path)
			return false
		}

		fmt.Printf("Error validating target path %s: %s\n", path, err)
		return false
	}

	if fileInfo.IsDir() {
		return true
	} else {
		fmt.Printf("%s is not a directory!\n", path)
	}

	mode := fileInfo.Mode()

	// Check for user read permission (0400) and user write permission (0200)
	userCanRead := mode&0400 != 0
	userCanWrite := mode&0200 != 0

	if userCanRead && userCanWrite {
		return true
	}

	if !userCanRead && !userCanWrite {
		fmt.Printf("user lacks both read and write permissions for: %s", path)
		return false
	} else if !userCanRead {
		fmt.Printf("user lacks read permission for: %s", path)
		return false
	} else {
		fmt.Printf("user lacks write permission for: %s", path)
		return false
	}
}

func createFileWithParents(targetFile string) (*os.File, error) {
	dir := filepath.Dir(targetFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error creating directory %s: %s", dir, err)
	}

	file, err := os.Create(targetFile)
	if err != nil {
		return nil, fmt.Errorf("unable to create local file %s", targetFile)
	}

	return file, nil
}
