package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Commit = "undefined"
var BuildTime = "undefined"

// TODO: use `rootCmd.Version` instead`
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of sandbox-argocd",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`{"commit": "%s", "build_time":"%s"}`, Commit, BuildTime)
		},
	}
}
