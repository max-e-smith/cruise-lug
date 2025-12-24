package cruise

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/max-e-smith/cruise-lug/cmd/common"
	"github.com/max-e-smith/cruise-lug/cmd/get"
	"github.com/max-e-smith/cruise-lug/cmd/get/cruise/dcdb"
	"github.com/spf13/cobra"
	"log"
	"path"
	"sync"
	"time"
)

var multibeam bool
var crowdsourced bool
var wcd bool
var trackline bool

var s3client s3.Client
var bucket = "noaa-dcdb-bathymetry-pds" // noaa-dcdb-bathymetry-pds.s3.amazonaws.com/index.html

var cruiseCmd = &cobra.Command{
	Use:   "cruise",
	Short: "Download NOAA survey data to local path",
	Long: `Use 'clug get cruise <survey(s)> <local path> <options>' to download marine geophysics data to your machine. 

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

		if !multibeam && !wcd && !trackline {
			fmt.Println("Please specify data type(s) for download.")
			fmt.Println(cmd.UsageString())
			return
		}

		download(surveys, path)

		fmt.Println("Done.")
	},
}

func init() {
	get.GetCmd.AddCommand(cruiseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cruiseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cruiseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Local flags
	cruiseCmd.Flags().BoolVarP(&multibeam, "multibeam-bathy", "m", false, "Download multibeam bathy data")
	cruiseCmd.Flags().BoolVarP(&crowdsourced, "crowdsourced-bathy", "c", false, "Download crowdsourced bathy data")
	cruiseCmd.Flags().BoolVarP(&wcd, "water-column", "w", false, "Download water column data")
	cruiseCmd.Flags().BoolVarP(&trackline, "trackline", "t", false, "Download trackline data")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		config.WithRegion("us-east-1"),
	)

	if err != nil {
		fmt.Printf("Error loading AWS config: %s\n", err)
		fmt.Println("Failed to download multibeam surveys.")
		return
	}

	s3client = *s3.NewFromConfig(cfg)
}

func download(surveys []string, targetPath string) {
	if !common.VerifyTarget(targetPath) {
		fmt.Printf("Quitting.")
		return
	}

	if multibeam {
		downloadBathySurveys(surveys, targetPath)
	}

	if crowdsourced {
	} // TODO

	if wcd {
	} // TODO

	if trackline {
	} // TODO

	return
}

func diskSpaceCheck(rootPaths []string, targetPath string) bool {
	availableSpace := common.GetAvailableDiskSpace(targetPath)
	totalSurveysSize, err := common.GetDiskUsageEstimate(bucket, s3client, rootPaths)

	if err != nil {
		log.Fatal(err)
		return false
	}

	if totalSurveysSize < 0 {
		totalSurveysSize = 0
	}

	fmt.Printf("  total download size: %gGB\n", common.ByteToGB(totalSurveysSize))
	fmt.Printf("  disk space available: %gGB\n", common.ByteToGB(int64(availableSpace)))

	if availableSpace > uint64(totalSurveysSize) {
		return true
	}

	return false
}

func logDownloadTime(start time.Time) {
	fmt.Printf("Download completed in %g hours.\n", common.HoursSince(start))
}

func downloadBathySurveys(surveys []string, targetPath string) {
	start := time.Now()
	defer logDownloadTime(start)

	fmt.Println("Resolving bathymetry data for specified surveys: ", surveys)
	var surveyRoots = dcdb.ResolveMultibeamSurveys(surveys, bucket, s3client)

	if len(surveyRoots) == 0 {
		fmt.Println("No surveys found.")
		return
	} else {
		fmt.Printf("Found %d of %d wanted surveys at: %s\n", len(surveyRoots), len(surveys), surveyRoots)
		// TODO additional verification of survey match results
	}

	fmt.Println("Checking available disk space")
	if !diskSpaceCheck(surveyRoots, targetPath) {
		fmt.Println("Specified path does not have enough disk space available.")
		return
	}

	fmt.Printf("Downloading survey files to %s...\n", targetPath)
	downloadFiles(surveyRoots, targetPath)

	fmt.Println("bathymetry data downloaded.")
}

func downloadFiles(prefixes []string, targetDir string) {
	for _, survey := range prefixes {
		var fileDownloadPageSize int32 = 10

		params := &s3.ListObjectsV2Input{
			Bucket:  aws.String(bucket),
			Prefix:  aws.String(survey),
			MaxKeys: aws.Int32(fileDownloadPageSize),
		}

		filePaginator := s3.NewListObjectsV2Paginator(&s3client, params)
		for filePaginator.HasMorePages() {
			page, err := filePaginator.NextPage(context.TODO())
			if err != nil {
				log.Fatal(err)
				return
			}

			var wg sync.WaitGroup
			for _, object := range page.Contents {
				wg.Add(1)
				go common.DownloadLargeObject(bucket, *object.Key, s3client, path.Join(targetDir, *object.Key), &wg)
			}
			wg.Wait()
		}
	}
}
