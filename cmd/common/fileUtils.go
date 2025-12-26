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

func getAvailableDiskSpace(localPath string) uint64 {
	usage := du.NewDiskUsage(localPath)
	if usage == nil {
		log.Fatalf("Could not get disk usage for path: %s", localPath)
	}
	return usage.Available() // bytes
}

func VerifyTarget(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("Target download path %s does not exist.\n", path)
		}
		return fmt.Errorf("Error validating target path %s: %s\n", path, err)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("%s is not a directory!\n", path)
	}

	mode := fileInfo.Mode()

	// Check for user read permission (0400) and user write permission (0200)
	userCanRead := mode&0400 != 0
	userCanWrite := mode&0200 != 0

	if userCanRead && userCanWrite {
		return nil
	}

	if !userCanRead && !userCanWrite {
		return fmt.Errorf("user lacks both read and write permissions for: %s", path)
	} else if !userCanRead {
		return fmt.Errorf("user lacks read permission for: %s", path)
	} else {
		return fmt.Errorf("user lacks write permission for: %s", path)
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

func DiskSpaceCheck(neededBytes int64, targetPath string) error {
	fmt.Println("Checking available disk space")
	if neededBytes < 0 {
		neededBytes = 0
	}

	availableSpace := getAvailableDiskSpace(targetPath)

	fmt.Printf("  total disk size needed: %gGB\n", ByteToGB(neededBytes))
	fmt.Printf("  disk space available: %gGB\n", ByteToGB(int64(availableSpace)))

	if availableSpace > uint64(neededBytes) {
		return nil
	}

	return fmt.Errorf("not enough available space")
}
