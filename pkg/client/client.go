package client

import (
	"os"
	"path/filepath"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewFromConfig(kubeconfig string) (runtimeclient.Client, error) {
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
