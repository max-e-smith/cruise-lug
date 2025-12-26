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
	"github.com/max-e-smith/cruise-lug/cmd/gets/dcdb"
	"github.com/spf13/cobra"
)

var multibeam bool
var crowdsourced bool
var wcd bool
var trackline bool

var s3client s3.Client

type CruiseRequest interface {
	ResolveSurveys() ([]string, error)
	CheckDiskAvailability() error
	DownloadSurveys() error
}

var cruiseCmd = &cobra.Command{
	Use:   "cruise",
	Short: "Download NOAA survey data to local path",
	Long: `Use 'clug get cruise <survey(s)> <local path> <options>' to download marine geophysics data to your machine. 

		Data is downloaded from the NOAA Open Data Dissemination cloud buckets by default. You must 
		specify a data type(s) for this command. View the help for more info on those options. Specify
		the survey(s) you want to download and a local path to download data to. The path must exist and 
		have the necessary permissions.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetPath, surveys, argsErr := getArgs(args)
		if argsErr != nil {
			usageError(cmd, argsErr)
			return
		}

		err := verifyUsage(targetPath)
		if err != nil {
			usageError(cmd, err)
			return
		}

		if multibeam {
			request := dcdb.MultibeamRequest{
				Surveys:     surveys,
				S3Client:    s3client,
				TargetDir:   targetPath,
				WorkerCount: 0,
			}
			if err := requestMultibeamDownload(request); err != nil {
				fmt.Println(err)
				return
			}
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

func usageError(cmd *cobra.Command, err error) {
	fmt.Println(err)
	fmt.Println(cmd.UsageString())
}

func getArgs(args []string) (string, []string, error) {
	var length = len(args)
	if length <= 1 {
		return "", nil, errors.New("please specify survey name(s) and a target file path")
	}

	var targetPath = args[length-1]
	var surveys = args[:length-1]

	return targetPath, surveys, nil
}

func verifyUsage(targetPath string) error {
	if !multibeam && !wcd && !trackline {
		return fmt.Errorf("please specify data type(s) for download")
	}

	targetError := common.VerifyTarget(targetPath)
	if targetError != nil {
		return targetError
	}
	return nil
}

func requestMultibeamDownload(request dcdb.MultibeamRequest) error {
	if surveys, resolveErr := request.ResolveSurveys(); resolveErr != nil {
		return resolveErr
	} else {
		request.Prefixes = surveys
	}

	if spaceErr := request.CheckDiskAvailability(); spaceErr != nil {
		return spaceErr
	}

	return request.DownloadSurveys()
}
