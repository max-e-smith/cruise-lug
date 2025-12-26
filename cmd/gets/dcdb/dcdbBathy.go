package dcdb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/max-e-smith/cruise-lug/cmd/common"
	"log"
	"path"
	"strings"
	"time"
)

var Bucket = "noaa-dcdb-bathymetry-pds" // https://noaa-dcdb-bathymetry-pds.s3.amazonaws.com/index.html

type MultibeamRequest struct {
	Surveys     []string
	Prefixes    []string
	S3Client    s3.Client
	TargetDir   string
	WorkerCount int
}

func logDownloadTime(start time.Time) {
	fmt.Printf("Download completed in %g hours.\n", common.HoursSince(start))
}

func (request MultibeamRequest) ResolveSurveys() ([]string, error) {
	fmt.Println("Resolving bathymetry data for specified surveys: ", request.Surveys)
	var surveyPaths []string
	wantedSurveys := len(request.Surveys)
	foundSurveys := 0

	pt, ptErr := request.S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:    aws.String(Bucket),
		Prefix:    aws.String("mb/"),
		Delimiter: aws.String("/"),
	})

	if ptErr != nil {
		return surveyPaths, ptErr
	}

	for _, platformType := range pt.CommonPrefixes {

		platformParams := &s3.ListObjectsV2Input{
			Bucket:    aws.String(Bucket),
			Prefix:    aws.String(*platformType.Prefix),
			Delimiter: aws.String("/"),
		}

		allPlatforms := s3.NewListObjectsV2Paginator(&request.S3Client, platformParams)

		for allPlatforms.HasMorePages() {
			platsPage, platsErr := allPlatforms.NextPage(context.TODO())

			if platsErr != nil {
				log.Fatal(platsErr)
				return surveyPaths, platsErr
			}
			for _, platform := range platsPage.CommonPrefixes {
				fmt.Printf("  searching %s\n", *platform.Prefix)

				platformParams := &s3.ListObjectsV2Input{
					Bucket:    aws.String(Bucket),
					Prefix:    aws.String(*platform.Prefix),
					Delimiter: aws.String("/"),
				}

				platformPaginator := s3.NewListObjectsV2Paginator(&request.S3Client, platformParams)

				for platformPaginator.HasMorePages() {
					surveysPage, err := platformPaginator.NextPage(context.TODO())
					if err != nil {
						return surveyPaths, err
					}

					for _, survey := range surveysPage.CommonPrefixes {
						surveyPrefix := *survey.Prefix
						survey := path.Base(strings.TrimRight(surveyPrefix, "/"))
						if isSurveyMatch(request.Surveys, survey) {
							surveyPaths = append(surveyPaths, surveyPrefix)
							foundSurveys++
						}
					}

				}
				if wantedSurveys == foundSurveys {
					// short circuit when enough surveys are found
					fmt.Printf("setting request prefixes to found surveys: %s\n", strings.Join(surveyPaths, ","))
					request.Prefixes = surveyPaths
					return surveyPaths, nil
				}
			}
		}
	}

	if len(surveyPaths) == 0 {
		return surveyPaths, fmt.Errorf("no surveys found")
	} else {
		// TODO additional verification of survey match results
	}
	fmt.Printf("Found %d of %d wanted surveys at: %s\n", len(surveyPaths), len(request.Surveys), surveyPaths)
	request.Prefixes = surveyPaths
	return surveyPaths, nil
}

func (request MultibeamRequest) CheckDiskAvailability() error {
	bytes, estimateErr := common.GetDiskUsageEstimate(Bucket, request.S3Client, request.Prefixes)
	if estimateErr != nil {
		return fmt.Errorf("unable to get disk usage estimate from s3 bucket: %w", estimateErr)
	}

	return common.DiskSpaceCheck(bytes, request.TargetDir)
}

func (request MultibeamRequest) DownloadSurveys() error {
	start := time.Now()
	defer logDownloadTime(start)

	order := common.Order{
		Bucket:      Bucket,
		Prefixes:    request.Prefixes,
		Client:      request.S3Client,
		TargetDir:   request.TargetDir,
		WorkerCount: request.WorkerCount,
	}
	return order.DownloadFiles()
}

func isSurveyMatch(surveys []string, resolvedSurvey string) bool {
	for _, survey := range surveys {
		if survey == resolvedSurvey {
			fmt.Println("Found matching survey: ", survey)
			return true
		}
	}
	return false
}
