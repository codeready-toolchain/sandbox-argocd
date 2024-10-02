package validation

import (
	"bytes"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	kbuild "sigs.k8s.io/kustomize/kustomize/v5/commands/build"
	kfsys "sigs.k8s.io/kustomize/kyaml/filesys"
)

func lookupKustomizationFile(logger *log.Logger, afs afero.Afero, componentPath string) (string, bool) {
	for _, k := range []string{"kustomization.yaml", "kustomization.yml", "Kustomization"} {
		p := filepath.Join(componentPath, k)
		if _, err := afs.Open(p); err == nil {
			logger.Debug("found Kustomization file", "path", p)
			return p, true
		}
	}
	return "", false
}

// Verifies that `kustomize build` completes successfully
func checkBuild(logger *log.Logger, fsys kfsys.FileSystem, path string) error {
	logger.Debug("👀 checking kustomize build", "path", path)
	buffy := new(bytes.Buffer)

	kcmd := kbuild.NewCmdBuild(fsys, &kbuild.Help{}, buffy)
	if err := kcmd.RunE(kcmd, []string{path}); err != nil {
		return err
	}
	return nil
}
