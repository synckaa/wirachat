package cmd

import (
	"wirachat/internal"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "create a chatting server",
	Long:  "create a chatting server",
	Run: func(cmd *cobra.Command, args []string) {
		RunServer()
	},
}

var MyPort string
var MyEndpoint string

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&MyPort, "port", "p", "8080", "port to serve on")
	serveCmd.Flags().StringVarP(&MyEndpoint, "endpoint", "e", "/", "path to serve on")

}

func RunServer() {
	cfg := internal.Configuration{
		Port: MyPort,
		Path: MyEndpoint,
	}

	internal.RunServeCommand(cfg)
}
