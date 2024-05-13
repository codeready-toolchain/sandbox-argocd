package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sandbox-argocd",
	Short: "Manage Argo CD for Dev Sandbox",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var kubeconfig string
var verbose bool

func init() {
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file (default to $KUBECONFIG or $HOME/.kube/config)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(NewAddAppCmd())
	rootCmd.AddCommand(NewListAppsCmd())
}
