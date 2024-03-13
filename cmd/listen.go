package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"webhooker/server"
)

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "starts the webhook server",
	Long: `A longer description:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		certFile, _ := cmd.Flags().GetString("certfile")
		keyFile, _ := cmd.Flags().GetString("keyfile")
		listenAddress, _ := cmd.Flags().GetString("listen-address")
		dbUri, _ := cmd.Flags().GetString("dburi")
		mySrv := server.Server{}
		err := mySrv.Listen(listenAddress, certFile, keyFile, dbUri)
		if err != nil {
			log.Fatalf("error starting the webhook server: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(listenCmd)
	listenCmd.PersistentFlags().String("listen-address", ":8080", "listen address and port")
	listenCmd.PersistentFlags().String("certfile", "", "certificate file")
	listenCmd.PersistentFlags().String("keyfile", "", "private key file")
	listenCmd.PersistentFlags().String("dburi", "", "MongoDB database connection URI")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
