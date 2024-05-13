package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/codeready-toolchain/sandbox-argocd/pkg/applications"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewListAppsCmd() *cobra.Command {
	var pathToApps string

	cmd := &cobra.Command{
		Use:   "list-applications --apps=<path/to/apps>",
		Short: "List Applications and ApplicationSets in the given 'apps'",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.New(cmd.OutOrStdout())
			logger.SetLevel(log.InfoLevel)
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}
			afs := afero.Afero{
				Fs: afero.NewOsFs(),
			}
			apps, appsets, err := applications.ListApplications(logger, afs, pathToApps)
			if err != nil {
				return err
			}
			logger.Infof("Applications:    %s", join(apps...))
			logger.Infof("ApplicationSets: %s", join(appsets...))
			return nil
		},
	}
	cmd.Flags().StringVarP(&pathToApps, "apps", "a", "", "Path to ArgoCD Application and ApplicationSets")
	if err := cmd.MarkFlagRequired("apps"); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return cmd

}

func join[T runtimeclient.Object](objs ...T) string {
	if len(objs) > 0 {

		buffy := &strings.Builder{}
		for i, obj := range objs {
			buffy.WriteString(obj.GetName())
			if i < len(objs)-1 {
				buffy.WriteString(", ")
			}
		}
		return buffy.String()
	}
	return "<none>"
}
