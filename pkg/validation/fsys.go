package validation

import (
	iofs "io/fs"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	kfsys "sigs.k8s.io/kustomize/kyaml/filesys"
)

func NewInMemoryFS(logger *log.Logger, afs afero.Afero, baseDir string) (kfsys.FileSystem, error) {
	fsys := kfsys.MakeFsInMemory()
	if err := afs.Walk(baseDir,
		func(path string, info iofs.FileInfo, err error) error {
			if err != nil {
				logger.Error("prevent panic by handling failure", "path", path, "err", err)
				return err
			}
			if info.IsDir() {
				logger.Debug("adding directory in fsys", "path", path)
				return fsys.Mkdir(path)
			}
			if ext := strings.ToLower(filepath.Ext(path)); ext == ".md" || ext == ".adoc" {
				logger.Debug("skipping file", "path", path)
				return nil
			}
			data, err := afs.ReadFile(path)
			if err != nil {
				return err
			}
			logger.Debug("adding file in fsys", "path", path)
			return fsys.WriteFile(path, data)
		},
	); err != nil {
		return nil, err
	}
	return fsys, nil
}
