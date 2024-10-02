package validation

import (
	"fmt"
	iofs "io/fs"
	"path/filepath"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Look for all YAML files in the given paths and when the contents if an Argo CD Application or ApplicationSet,
// verify that the `spec.source.path` matches an existing component
func CheckApplications(logger *log.Logger, afs afero.Afero, baseDir string, apps ...string) error {

	for _, path := range apps {
		p := filepath.Join(baseDir, path)
		logger.Info("👀 checking Applications and ApplicationSets", "path", path)
		fsys, err := NewInMemoryFS(logger, afs, p)
		if err != nil {
			return err
		}
		if err := afs.Walk(p, func(path string, info iofs.FileInfo, err error) error {
			if err != nil {
				logger.Error("prevent panic by handling failure", "path", path)
				return err
			}
			if info.IsDir() {
				logger.Debug("👀 checking contents", "path", path)
				if kpath, found := lookupKustomizationFile(logger, afs, path); found {
					if err := checkKustomizeResources(logger, afs, baseDir, kpath); err != nil {
						return err
					}
					if info.Name() != "base" {
						if err := checkBuild(logger, fsys, path); err != nil {
							return err
						}
					}
				}
				return nil
			}
			data, err := afs.ReadFile(path)
			if err != nil {
				return err
			}
			if filepath.Ext(info.Name()) == ".yaml" {
				logger.Debug("checking contents", "path", path)
				app := &argocdv1alpha1.Application{}
				if err := yaml.Unmarshal(data, app); err == nil && app.Spec.Source != nil {
					return checkPath(afs, baseDir, app.Spec.Source.Path)
				}
				appSet := &argocdv1alpha1.ApplicationSet{}
				if err := yaml.Unmarshal(data, appSet); err == nil && appSet.Spec.Template.Spec.Source != nil {
					return checkPath(afs, baseDir, appSet.Spec.Template.Spec.Source.Path)
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func checkPath(afs afero.Afero, repoURL, path string) error {
	p := filepath.Join(repoURL, path)
	if _, err := afs.ReadDir(p); err != nil {
		return fmt.Errorf("%s is not valid", path)
	}
	// also, check that the path contains a `kustomization.yaml` file
	if exists, err := afs.Exists(filepath.Join(p, "kustomization.yaml")); err != nil || !exists {
		return fmt.Errorf("%s does not contain a 'kustomization.yaml' file", path)
	}

	return nil
}
