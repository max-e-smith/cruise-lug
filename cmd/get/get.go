package get

import (
	"github.com/max-e-smith/cruise-lug/cmd"
	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Download NOAA survey data to local path",
	Long:  `Use 'clug get <subcommand> to download a dataset from the Noaa Open Data Dissemination cloud.`,
	//Run: func(cmd *cobra.Command, args []string) {
	//	var length = len(args)
	//	if length <= 1 {
	//		fmt.Println("Please specify a subcommand")
	//		fmt.Println(cmd.UsageString())
	//		return
	//	}
	//},
}

func init() {
	cmd.rootCmd.AddCommand(GetCmd)
}
