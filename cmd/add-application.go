package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/charmbracelet/log"
	"github.com/codeready-toolchain/sandbox-argocd/pkg/apps"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {

}

func NewAddAppCmd() *cobra.Command {
	var pathToApps string
	var repositoryURL string
	var targetRevision string

	cmd := &cobra.Command{
		Use:   "add-application <name> --apps=<path/to/apps> --repo-url=<url> --target-revision=<revision> --kubeconfig=<path/to/kubeconfig>",
		Short: "Add an Application or ApplicationSet",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.New(cmd.OutOrStdout())
			logger.SetLevel(log.InfoLevel)
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}
			cl, err := newClientFromConfig(kubeconfig)
			if err != nil {
				logger.Errorf(err.Error())
				os.Exit(1)
			}
			afs := afero.Afero{
				Fs: afero.NewOsFs(),
			}
			apps, appsets, err := apps.LookupApplications(logger, afs, pathToApps)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}
			for _, app := range apps {
				if app.Name == args[0] {
					app.Spec.Source.RepoURL = repositoryURL
					app.Spec.Source.TargetRevision = targetRevision
					existingApp := &argocdv1alpha1.Application{}
					if err := cl.Get(cmd.Context(), types.NamespacedName{
						Namespace: app.Namespace,
						Name:      app.Name,
					}, existingApp); err == nil {
						// app already exist, let's update it instead
						app.ResourceVersion = existingApp.ResourceVersion
						if err := cl.Update(cmd.Context(), app, &runtimeclient.UpdateOptions{}); err != nil {
							logger.Error(err.Error())
							os.Exit(1)
						}
						logger.Infof("successfully updated the '%s' Argo CD Application", app.Name)
						os.Exit(0)
					}
					if err := cl.Create(cmd.Context(), app, &runtimeclient.CreateOptions{}); err != nil {
						logger.Error(err.Error())
						os.Exit(1)
					}
					logger.Infof("successfully created the '%s' Argo CD Application", app.Name)
					os.Exit(0)
				}
			}

			for _, appset := range appsets {
				if appset.Name == args[0] {
					appset.Spec.Template.Spec.Source.RepoURL = repositoryURL
					appset.Spec.Template.Spec.Source.TargetRevision = targetRevision
					existingAppset := &argocdv1alpha1.ApplicationSet{}
					if err := cl.Get(cmd.Context(), types.NamespacedName{
						Namespace: appset.Namespace,
						Name:      appset.Name,
					}, existingAppset); err == nil {
						// app already exist, let's update it instead
						appset.ResourceVersion = existingAppset.ResourceVersion
						if err := cl.Update(cmd.Context(), appset, &runtimeclient.UpdateOptions{}); err != nil {
							logger.Error(err.Error())
							os.Exit(1)
						}
						logger.Infof("successfully updated the '%s' Argo CD ApplicationSet", appset.Name)
						os.Exit(0)
					}
					if err := cl.Create(cmd.Context(), appset, &runtimeclient.CreateOptions{}); err != nil {
						logger.Error(err.Error())
						os.Exit(1)
					}
					logger.Infof("successfully created the '%s' Argo CD ApplicationSet", appset.Name)
					os.Exit(0)
				}
			}
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

func newClientFromConfig(kubeconfig string) (runtimeclient.Client, error) {
	d, err := os.ReadFile(locateKubeconfig(kubeconfig))
	if err != nil {
		return nil, err
	}
	clientCfg, err := clientcmd.NewClientConfigFromBytes(d)
	if err != nil {
		return nil, err
	}
	cfg, err := clientCfg.ClientConfig()
	if err != nil {
		return nil, err
	}
	cfg.APIPath = "/api"
	cfg.GroupVersion = &schema.GroupVersion{Version: "v1"}
	cfg.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	s := scheme.Scheme
	if err := argocdv1alpha1.AddToScheme(s); err != nil {
		return nil, err
	}
	return runtimeclient.New(cfg, runtimeclient.Options{
		Scheme: s,
	})
}

// locateKubeconfig returns a file reader on (by order of match):
// - the --kubeconfig CLI argument if it was provided
// - the $KUBECONFIG file it the env var was set
// - the <user_home_dir>/.kube/config file
func locateKubeconfig(kubeconfig string) string {
	var path string
	if kubeconfig != "" {
		path = kubeconfig
	} else if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig != "" {
		path = kubeconfig
	} else {
		path = filepath.Join(homeDir(), ".kube", "config")
	}
	return path
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
