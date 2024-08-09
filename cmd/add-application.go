package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/codeready-toolchain/sandbox-argocd/pkg/applications"
	"github.com/codeready-toolchain/sandbox-argocd/pkg/client"

	"github.com/agnivade/levenshtein"
	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewAddAppCmd() *cobra.Command {
	var pathToApps string
	var repositoryURL string
	var targetRevision string

	cmd := &cobra.Command{
		Use:   "add-application <name> --apps=<path/to/apps> --repo-url=<url> --target-revision=<revision> --kubeconfig=<path/to/kubeconfig>",
		Short: "Add an Application or ApplicationSet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.New(cmd.OutOrStdout())
			logger.SetLevel(log.InfoLevel)
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}
			cl, err := client.NewFromConfig(kubeconfig)
			if err != nil {
				logger.Errorf(err.Error())
				os.Exit(1)
			}
			afs := afero.Afero{
				Fs: afero.NewOsFs(),
			}
			apps, appsets, err := applications.ListApplications(logger, afs, pathToApps)
			if err != nil {
				return err
			}
			for _, app := range apps {
				if app.Name == args[0] {
					app.Spec.Source.RepoURL = repositoryURL
					app.Spec.Source.TargetRevision = targetRevision
					return applications.CreateApplication(cmd.Context(), logger, cl, app)
				}
			}

			for _, appset := range appsets {
				if appset.Name == args[0] {
					appset.Spec.Template.Spec.Source.RepoURL = repositoryURL
					appset.Spec.Template.Spec.Source.TargetRevision = targetRevision
					return applications.CreateApplicationSet(cmd.Context(), logger, cl, appset)
				}
			}
			logger.Errorf("🤷 unable to find the '%s' Argo CD Application/ApplicationSet", args[0])

			// in this case, suggest the closest apps/appsets
			suggestions := []string{}
			threshold := 4
			for _, app := range apps {
				if distance := levenshtein.ComputeDistance(args[0], app.Name); distance < threshold {
					suggestions = append(suggestions, app.Name)
				}
			}
			for _, appset := range appsets {
				if distance := levenshtein.ComputeDistance(args[0], appset.Name); distance < threshold {
					suggestions = append(suggestions, appset.Name)
				}
			}
			if len(suggestions) > 0 {
				logger.Infof("🤔 did you mean: %s", strings.Join(suggestions, ", "))
			} else {
				logger.Info("🤨 no similar Application or ApplicationSet")
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&pathToApps, "apps", "a", "", "Path to ArgoCD Application and ApplicationSets")
	if err := cmd.MarkFlagRequired("apps"); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	cmd.Flags().StringVar(&repositoryURL, "repo-url", "", "Application's Repository URL (overridding the .spec value)")
	if err := cmd.MarkFlagRequired("repo-url"); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	cmd.Flags().StringVar(&targetRevision, "target-revision", "", "Application's Target revision (overridding the .spec value)")
	if err := cmd.MarkFlagRequired("target-revision"); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return cmd
}
