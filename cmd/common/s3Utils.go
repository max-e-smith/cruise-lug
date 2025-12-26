package common

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"path"
	"sync"
	"time"
)

func GetDiskUsageEstimate(bucket string, s3client s3.Client, rootPaths []string) (int64, error) {
	var totalSurveysSize int64 = 0

	fmt.Printf("Getting disk usage estimate from s3 bucket: %s\n", bucket)
	fmt.Printf("Getting disk usage estimate for root paths: %s\n", rootPaths)
	for _, surveyRootPath := range rootPaths {
		fmt.Printf("Getting disk usage estimate for %s\n", surveyRootPath)
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

type Order struct {
	Bucket      string
	Prefixes    []string
	Client      s3.Client
	TargetDir   string
	WorkerCount int
}

func (manifest Order) DownloadFiles() error {
	fmt.Printf("Downloading files to %s...\n", manifest.TargetDir)
	for _, survey := range manifest.Prefixes {
		var fileDownloadPageSize int32 = 10

		params := &s3.ListObjectsV2Input{
			Bucket:  aws.String(manifest.Bucket),
			Prefix:  aws.String(survey),
			MaxKeys: aws.Int32(fileDownloadPageSize),
		}

		filePaginator := s3.NewListObjectsV2Paginator(&manifest.Client, params)
		for filePaginator.HasMorePages() {
			page, err := filePaginator.NextPage(context.TODO())
			if err != nil {
				return err
			}

			var wg sync.WaitGroup
			for _, object := range page.Contents {
				wg.Add(1)
				go downloadLargeObject(manifest.Bucket, *object.Key, manifest.Client, path.Join(manifest.TargetDir, *object.Key), &wg)
			}
			wg.Wait()
		}
		fmt.Println("files downloaded.")
	}
	return nil
}

func downloadLargeObject(bucket string, objectKey string, client s3.Client, targetFile string, wg *sync.WaitGroup) {
	defer wg.Done()
	//fmt.Printf("Downloading %s to %s...\n", objectKey, bucket)

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
