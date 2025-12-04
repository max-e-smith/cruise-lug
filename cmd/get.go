package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var bathy bool
var wcd bool
var trackline bool

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Download NOAA survey data to local path",
	Long: `Use 'clug get <survey(s)> <local path> <options>' to download marine geophysics data to your machine. 

		Data is downloaded from the NOAA Open Data Dissemination cloud bucket by default. Use
		the config command to view and change buckets used for each data type. You must specify a 
		data type(s) to download by specifying one or more of the data options for this command. 
		View the help for this command for more info on those options.`,
	Run: func(cmd *cobra.Command, args []string) {

		var length = len(args)
		if length <= 1 {
			fmt.Println("Please specify survey name(s) and a target file path.")
			fmt.Println(cmd.UsageString())
		}

		var path = args[length-1]
		var surveys = args[:length-1] // high is non-inclusive

		if !bathy || !wcd || !trackline {
			fmt.Println("Please specify data type(s) for download.")
			fmt.Println(cmd.UsageString())
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

func download(surveys []string, path string) {
	if bathy {
		fmt.Println("resolving bathymetry data for provided surveys")
		// TODO resolve surveys
		fmt.Println("checking available disk space")
		// TODO verify survey size against available disk space and report out
		fmt.Println("downloading surveys ", surveys, " to ", path, "...")
		// TODO recursively download surveys
		fmt.Println("bathymetry data downloaded.")
	}

	if wcd {
		fmt.Println("resolving water column data for provided surveys")
		// TODO resolve surveys
		fmt.Println("checking available disk space")
		// TODO verify survey size against available disk space and report out
		fmt.Println("downloading surveys ", surveys, " to ", path, "...")
		// TODO recursively download surveys
		fmt.Println("bathymetry data downloaded.")
	}

	if trackline {
		fmt.Println("resolving bathymetry data for provided surveys")
		// TODO resolve surveys
		fmt.Println("checking available disk space")
		// TODO verify survey size against available disk space and report out
		fmt.Println("downloading surveys ", surveys, " to ", path, "...")
		// TODO recursively download surveys
		fmt.Println("bathymetry data downloaded.")
	}

	return
}

func resolveBathySurveys() {
	// TODO
}

func resolveWaterColumnSurveys() {
	// TODO
}

func resolveTracklineSurveys() {
	// TODO
}
