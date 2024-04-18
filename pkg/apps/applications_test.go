package apps_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/codeready-toolchain/sandbox-argocd/pkg/apps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupApplications(t *testing.T) {

	baseDir := "/path/to/apps"

	t.Run("empty dir", func(t *testing.T) {
		// given
		afs, err := newAfs(baseDir)
		require.NoError(t, err)
		logger := log.New(os.Stdout)

		// when
		apps, appsets, err := apps.LookupApplications(logger, afs, baseDir)

		// then
		require.NoError(t, err)
		require.Empty(t, apps)
		require.Empty(t, appsets)
	})

	t.Run("1 app and 1 appset", func(t *testing.T) {
		// given
		afs, err := newAfs(baseDir)
		require.NoError(t, err)
		err = afs.WriteFile(filepath.Join(baseDir, "app-cookie.yaml"), appCookieData, 0755)
		require.NoError(t, err)
		err = afs.WriteFile(filepath.Join(baseDir, "appset-pasta.yaml"), appsetPastaData, 0755)
		require.NoError(t, err)
		logger := log.New(os.Stdout)

		// when
		apps, appsets, err := apps.LookupApplications(logger, afs, baseDir)

		// then
		require.NoError(t, err)
		require.Len(t, apps, 1)
		assert.Equal(t, "app-cookie", apps[0].Name)
		require.Len(t, appsets, 1)
		assert.Equal(t, "appset-pasta", appsets[0].Name)
	})
}

var appCookieData = []byte(`
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: app-cookie
spec:
  destination:
    server: https://kubernetes.default.svc
  project: default
  source:
    path: components/cookie`)

var appsetPastaData = []byte(`
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: appset-pasta
spec:
  template:
    spec:
      destination:
        server: "{{server}}"
      project: default
      source:
        path: components/pasta`)

func newAfs(baseDir string) (afero.Afero, error) {
	afs := afero.Afero{
		Fs: afero.NewMemMapFs(),
	}
	err := afs.Mkdir(baseDir, os.ModeDir)
	return afs, err
}
