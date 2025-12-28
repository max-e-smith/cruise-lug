package get

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/max-e-smith/cruise-lug/cmd"
	"github.com/max-e-smith/cruise-lug/cmd/common"
	"github.com/max-e-smith/cruise-lug/cmd/gets/cruise"
	"github.com/spf13/cobra"
	"log"
)

var multibeam bool
var crowdsourced bool
var wcd bool
var trackline bool

var s3client s3.Client

var cruiseCmd = &cobra.Command{
	Use:   "cruise",
	Short: "Download NOAA survey data to local path",
	Long: `Use 'clug get cruise <survey(s)> <local path> <options>' to download marine geophysics data to your machine. 

		Data is downloaded from the NOAA Open Data Dissemination cloud buckets by default. You must 
		specify a data type(s) for this command. View the help for more info on those options. Specify
		the survey(s) you want to download and a local path to download data to. The path must exist and 
		have the necessary permissions.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetPath, surveys := parseArgs(cmd, args)

		if multibeam {
			requestMultibeamDownload(
				dcdb.MultibeamRequest{
					Surveys:     surveys,
					S3Client:    s3client,
					TargetDir:   targetPath,
					WorkerCount: 0,
				},
			)
		}

		if crowdsourced {
		} // TODO

		if wcd {
		} // TODO

		if trackline {
		} // TODO

		fmt.Println("Done.")
		return
	},
}

func init() {
	cmd.GetCmd.AddCommand(cruiseCmd)

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

func parseArgs(cmd *cobra.Command, args []string) (string, []string) {
	var length = len(args)
	if length <= 1 {
		usageError(cmd, errors.New("please specify survey name(s) and a target file path"))
	}

	if !multibeam && !wcd && !trackline {
		usageError(cmd, errors.New("please specify data type(s) for download"))
	}

	var targetPath = args[length-1]
	var surveys = args[:length-1]

	targetError := common.VerifyTarget(targetPath)
	if targetError != nil {
		usageError(cmd, targetError)
	}

	return targetPath, surveys
}

func usageError(cmd *cobra.Command, err error) {
	fmt.Println(cmd.UsageString())
	log.Fatal(err)
}

func requestMultibeamDownload(request dcdb.MultibeamRequest) {

	request.ResolveSurveys()
	request.CheckDiskAvailability()
	request.DownloadSurveys()

	if request.Error != nil {
		log.Fatal(request.Error)
	}
}
