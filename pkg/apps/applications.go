package apps

import (
	fs "io/fs"
	"path/filepath"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

func LookupApplications(logger *log.Logger, afs afero.Afero, baseDir string) ([]*argocdv1alpha1.Application, []*argocdv1alpha1.ApplicationSet, error) {
	logger.Info("👀 looking for Applications", "path", baseDir)
	apps := []*argocdv1alpha1.Application{}
	appsets := []*argocdv1alpha1.ApplicationSet{}
	err := afs.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			logger.Error("prevent panic by handling failure", "path", path)
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := afs.ReadFile(path)
		if err != nil {
			return err
		}
		if filepath.Ext(info.Name()) == ".yaml" {
			logger.Debug("checking contents", "path", path)
			app := &argocdv1alpha1.Application{}
			if err := yaml.Unmarshal(data, app); err == nil && app.Spec.Destination.Server != "" {
				apps = append(apps, app)
			}
			appset := &argocdv1alpha1.ApplicationSet{}
			if err := yaml.Unmarshal(data, appset); err == nil && appset.Spec.Template.Spec.Destination.Server != "" {
				appsets = append(appsets, appset)
			}
		}
		return nil
	})
	return apps, appsets, err
}
