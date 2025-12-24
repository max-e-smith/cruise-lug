package common

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"sync"
	"time"
)

func GetDiskUsageEstimate(bucket string, s3client s3.Client, rootPaths []string) (int64, error) {
	var totalSurveysSize int64 = 0

	for _, surveyRootPath := range rootPaths {
		// TODO paginate
		result, err := s3client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
			Prefix: aws.String(surveyRootPath),
		})

		if err != nil {
			return totalSurveysSize, err
		}

		for _, object := range result.Contents {
			//log.Printf("key=%s size=%d", aws.ToString(object.Key), *object.Size)
			totalSurveysSize = totalSurveysSize + *object.Size
		}
	}

	return totalSurveysSize, nil
}

func DownloadLargeObject(bucket string, objectKey string, client s3.Client, targetFile string, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := createFileWithParents(targetFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer closeFileChecked(file)

	start := time.Now()

	downloader := manager.NewDownloader(&client)
	n, err := downloader.Download(context.TODO(), file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		fmt.Printf("failed to download file: %w", err)
		return
	}

	fmt.Printf("Successfully downloaded %g GB to %s in %g minutes.\n", ByteToGB(n), targetFile, MinutesSince(start))
	return
}
