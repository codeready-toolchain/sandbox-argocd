package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/codeready-toolchain/sandbox-argocd/pkg/validation"

	charmlog "github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewValidateConfigCmd() *cobra.Command {

	var apps, components []string
	var baseDir string
	var verbose bool

	checkCmd := &cobra.Command{
		Use:   "check-config --base-dir=$(pwd) --apps apps-of-apps,apps --components components --verbose=false",
		Short: "Checks the Argo CD configuration",
		Args:  cobra.ExactArgs(0),

		Run: func(cmd *cobra.Command, _ []string) {
			logger := charmlog.New(cmd.OutOrStderr())
			logger.SetLevel(charmlog.InfoLevel)
			if verbose {
				logger.SetLevel(charmlog.DebugLevel)
			}
			logger.Info("🏁 Checking Argo CD configuration", "base-dir", baseDir)
			afs := afero.Afero{
				Fs: afero.NewOsFs(),
			}
			// verifies that the source path of the Applications and ApplicationSets exists
			if err := validation.CheckApplications(logger, afs, baseDir, apps...); err != nil {
				logger.Error(strings.ReplaceAll(err.Error(), ": ", ":\n"))
				os.Exit(1)
			}
			// verifies that `kustomize build` on each component completes successfully
			if err := validation.CheckComponents(logger, afs, baseDir, components...); err != nil {
				logger.Error(strings.ReplaceAll(err.Error(), ": ", ":\n"))
				os.Exit(1)
			}
		},
	}

	checkCmd.Flags().StringSliceVar(&apps, "apps", []string{}, "path(s) to the applications (comma-separated, relative to '--baseDir')")
	if err := checkCmd.MarkFlagRequired("apps"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %s", err))
	}
	checkCmd.Flags().StringVar(&baseDir, "base-dir", ".", "base directory of the repository")
	checkCmd.Flags().StringSliceVar(&components, "components", []string{}, "path(s) to the components (comma-separated, relative to '--baseDir')")
	if err := checkCmd.MarkFlagRequired("components"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %s", err))
	}
	checkCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	return checkCmd

}
