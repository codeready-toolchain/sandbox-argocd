package validation

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"sigs.k8s.io/kustomize/api/types"
)

// Compares the entries of `resources` in the Kustomize file with the contents in the current directory to see if
// any local file is missing (not referenced as a resource).
// Files starting with an underscore character (`_`) are ignored
func checkKustomizeResources(logger *log.Logger, afs afero.Afero, basedir, kpath string) error {
	logger.Debug("checking kustomization resource", "path", kpath)
	data, err := afs.ReadFile(kpath)
	if err != nil {
		return err
	}
	var kobj types.Kustomization
	if err := kobj.Unmarshal(data); err != nil {
		return err
	}

	// list resources
	logger.Debug("checking kustomization resources", "dir", filepath.Dir(kpath))
	entries, err := afs.ReadDir(filepath.Dir(kpath))
	if err != nil {
		return err
	}
entries:
	for _, e := range entries {
		name := e.Name()
		switch {
		case strings.HasPrefix(name, "_"):
			logger.Debug("ignoring file or dir prefix with underscore", "path", kpath)
			continue entries
		case name == filepath.Base(kpath):
			logger.Debug("ignoring base directory", "path", kpath)
			continue entries
		case !e.IsDir() && !(filepath.Ext(name) == ".yaml" || filepath.Ext(name) == ".yml"):
			logger.Debug("ignoring non-YAML file", "path", kpath)
			continue entries
		case filepath.Base(name) == "kustomization.yaml":
			logger.Debug("ignoring kustomization file", "path", kpath)
			continue entries
		}

		for _, r := range kobj.Resources {
			if filepath.Clean(r) == name {
				continue entries
			}
		}
		for _, sg := range kobj.ConfigMapGenerator {
			for _, f := range sg.FileSources {
				if i := strings.LastIndex(f, "="); i > 0 {
					if f[i+1:] == name {
						continue entries
					}
				} else if f == name {
					continue entries
				}
			}
		}
		for _, sg := range kobj.SecretGenerator {
			for _, f := range sg.FileSources {
				if i := strings.LastIndex(f, "="); i > 0 {
					if f[i+1:] == name {
						continue entries
					}
				} else if f == name {
					continue entries
				}
			}
		}
		for _, m := range kobj.PatchesStrategicMerge { //nolint:staticcheck
			if string(m) == name {
				continue entries
			}
		}
		for _, m := range kobj.Patches {
			if string(m.Path) == name {
				continue entries
			}
		}
		for _, m := range kobj.Transformers {
			if string(m) == name {
				continue entries
			}
		}
		rkpath, _ := filepath.Rel(basedir, kpath)

		return fmt.Errorf("resource is not referenced in %s: %s", rkpath, name)
	}
	return nil
}
