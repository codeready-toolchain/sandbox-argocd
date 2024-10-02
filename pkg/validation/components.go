package validation

import (
	"io/fs"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
)

// Looks for a `kustomization.yaml` file in all `components` directories and subdirs,
// and attempt to run `kustomize build`
func CheckComponents(logger *log.Logger, afs afero.Afero, baseDir string, components ...string) error {

	for _, path := range components {
		p := filepath.Join(baseDir, path)
		logger.Info("👀 checking components", "path", path)
		fsys, err := NewInMemoryFS(logger, afs, p)
		if err != nil {
			return err
		}
		if err := afs.Walk(p, func(path string, d fs.FileInfo, err error) error {
			if err != nil {
				logger.Error("prevent panic by handling failure", "path", path)
				return err
			}
			if !d.IsDir() {
				// skip
				return nil
			}
			// look for a Kustomization file in the directory
			if kp, found := lookupKustomizationFile(logger, afs, path); found {
				if err := checkKustomizeResources(logger, afs, baseDir, kp); err != nil {
					return err
				}
				if d.Name() != "base" {
					logger.Debug("checking Kustomization build ", "path", path)
					if err := checkBuild(logger, fsys, path); err != nil {
						return err
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
