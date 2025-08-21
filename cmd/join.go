package cmd

import (
	"wirachat/internal"

	"github.com/spf13/cobra"
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "join to chat server",
	Long:  "join to chat server",
	Run: func(cmd *cobra.Command, args []string) {
		JoinChat()
	},
}
var Myurl string

func init() {
	rootCmd.AddCommand(joinCmd)
	joinCmd.Flags().StringVarP(&Myurl, "url", "u", "ws://localhost:8080/", "url to serve on")

}

func JoinChat() {

	internal.RunChat(Myurl)
}
