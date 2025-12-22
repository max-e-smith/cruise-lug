package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "clug",
	Short: "A simple retrieval tool for ocean datasets hosted by the NODD",
	Long: `A CLI library for downloading ocean (bathymetry, trackline, and water column)
	data from the NOAA Open Data Dissemination (NODD) cloud on a survey by survey basis.
	This simplifies s3 object retrieval, which will almost always need to be downloaded 
	in batch groups, avoiding downloading each file object manually. 

	get, given a survey name argument or path, will download all survey files or 
	sub-survey files at a given path.

	glance will summarize all files and file sizes for an equivalent get command

	list will display all files that will be downloaded for an equivalent get command

	config can be used to change default bucket name and download parameters.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cruise-data-lug.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
