package cmd

import (
	"wirachat/internal"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "create a chatting server",
	Long:  "create a chatting server",
	Run: func(cmd *cobra.Command, args []string) {
		RunServer()
	},
}

var MyPort string
var MyPath string

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&MyPort, "port", "p", "8080", "port to serve on")
	serveCmd.Flags().StringVarP(&MyPath, "path", "e", "/", "path to serve on")

}

func RunServer() {
	cfg := internal.Configuration{
		Port: MyPort,
		Path: MyPath,
	}

	internal.RunServeCommand(cfg)
}
