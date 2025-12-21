package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
	"strings"
)

var bathy bool
var wcd bool
var trackline bool

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Download NOAA survey data to local path",
	Long: `Use 'clug get <survey(s)> <local path> <options>' to download marine geophysics data to your machine. 

		Data is downloaded from the NOAA Open Data Dissemination cloud buckets by default. You must 
		specify a data type(s) for this command. View the help for more info on those options. Specify
		the survey(s) you want to download and a local path to download data to. The path must exist and 
		have the necessary permissions.`,
	Run: func(cmd *cobra.Command, args []string) {
		var length = len(args)
		if length <= 1 {
			fmt.Println("Please specify survey name(s) and a target file path.")
			fmt.Println(cmd.UsageString())
			return
		}

		var path = args[length-1]
		var surveys = args[:length-1]

		if !bathy && !wcd && !trackline {
			fmt.Println("Please specify data type(s) for download.")
			fmt.Println(cmd.UsageString())
			return
		}

		download(surveys, path)

		fmt.Println("Done.")
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Local flags
	getCmd.Flags().BoolVarP(&bathy, "bathy", "b", false, "Download bathy data")
	getCmd.Flags().BoolVarP(&wcd, "water-column", "w", false, "Download water column data")
	getCmd.Flags().BoolVarP(&trackline, "trackline", "t", false, "Download trackline data")

}

func download(surveys []string, targetPath string) {
	if !verifyTarget(targetPath) {
		fmt.Printf("Quitting.")
		return
	}

	if bathy {
		downloadBathySurveys(surveys, targetPath)
	}

	if wcd {
		// TODO
	}

	if trackline {
		// TODO
	}

	return
}

func verifyTarget(path string) bool {
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

func diskSpaceCheck(rootPaths []string, targetPath string) bool {

	return false
}

func downloadBathySurveys(surveys []string, targetPath string) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		fmt.Printf("Error loading AWS config: %s\n", err)
		fmt.Println("Failed to download bathy surveys.")
		return
	}

	client := s3.NewFromConfig(cfg)

	// noaa-dcdb-bathymetry-pds.s3.amazonaws.com/index.html
	bucket := "noaa-dcdb-bathymetry-pds"

	fmt.Println("resolving bathymetry data for provided surveys: ", surveys)
	var surveyRoots []string = resolveBathySurveys(surveys, *client, bucket)

	if len(surveyRoots) == 0 {
		fmt.Println("No surveys found.")
		return
	} else {
		fmt.Printf("Found %d of %d wanted surveys at: %s\n", len(surveyRoots), len(surveys), surveyRoots)
		// TODO additional verification of survey match results
	}

	fmt.Println("checking available disk space")
	if !diskSpaceCheck(surveyRoots, targetPath) {
		fmt.Println("Specified path does not have enough disk space available.")
		return
	}

	fmt.Println("downloading surveys ", surveys, " to ", targetPath)
	// TODO recursively download surveys

	fmt.Println("bathymetry data downloaded.")
}

func resolveBathySurveys(inputSurveys []string, client s3.Client, bucket string) []string {
	var surveyPaths []string
	wantedSurveys := len(inputSurveys)
	foundSurveys := 0

	pt, ptErr := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String("mb/"),
		Delimiter: aws.String("/"),
	})

	if ptErr != nil {
		log.Fatal(ptErr)
		return surveyPaths
	}

	for _, platformType := range pt.CommonPrefixes {

		platformParams := &s3.ListObjectsV2Input{
			Bucket:    aws.String(bucket),
			Prefix:    aws.String(*platformType.Prefix),
			Delimiter: aws.String("/"),
		}

		allPlatforms := s3.NewListObjectsV2Paginator(&client, platformParams)

		for allPlatforms.HasMorePages() {
			platsPage, platsErr := allPlatforms.NextPage(context.TODO())

			if platsErr != nil {
				log.Fatal(platsErr)
				return []string{}
			}
			for _, platform := range platsPage.CommonPrefixes {
				fmt.Printf("  searching %s\n", *platform.Prefix)

				platformParams := &s3.ListObjectsV2Input{
					Bucket:    aws.String(bucket),
					Prefix:    aws.String(*platform.Prefix),
					Delimiter: aws.String("/"),
				}

				platformPaginator := s3.NewListObjectsV2Paginator(&client, platformParams)

				for platformPaginator.HasMorePages() {
					surveysPage, err := platformPaginator.NextPage(context.TODO())
					if err != nil {
						log.Fatal(err)
						return []string{}
					}

					for _, survey := range surveysPage.CommonPrefixes {
						surveyPrefix := *survey.Prefix
						survey := path.Base(strings.TrimRight(surveyPrefix, "/"))
						if isSurveyMatch(inputSurveys, survey) {
							surveyPaths = append(surveyPaths, surveyPrefix)
							foundSurveys++
						}
					}

				}
				if wantedSurveys == foundSurveys {
					return surveyPaths
				}
			}
		}
	}

	//for _, object := range result.Contents {
	//	log.Printf("key=%s size=%d", aws.ToString(object.Key), *object.Size)
	//}

	return surveyPaths
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
