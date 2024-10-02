package validation_test

import (
	"os"
	"testing"

	"github.com/codeready-toolchain/sandbox-argocd/pkg/validation"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemory(t *testing.T) {

	// given
	logger := log.New(os.Stdout)

	t.Run("valid filename", func(t *testing.T) {
		// given
		afs := afero.Afero{
			Fs: afero.NewMemMapFs(),
		}
		err := afs.MkdirAll("/basedir/apps", 0755)
		require.NoError(t, err)
		data := []byte("cookies are yummy")
		err = afs.WriteFile("/basedir/apps/kustomization.yaml", data, 0755)
		require.NoError(t, err)

		// when
		fsys, err := validation.NewInMemoryFS(logger, afs, "/basedir")

		// then
		require.NoError(t, err)

		assert.True(t, fsys.Exists("/basedir/apps"))
		assert.True(t, fsys.Exists("/basedir/apps/kustomization.yaml"))
		actual, err := fsys.ReadFile("/basedir/apps/kustomization.yaml")
		require.NoError(t, err)
		assert.Equal(t, data, actual)
	})

	t.Run("invalid filename", func(t *testing.T) {
		// given
		afs := afero.Afero{
			Fs: afero.NewMemMapFs(),
		}
		err := afs.MkdirAll("/basedir/apps", 0755)
		require.NoError(t, err)
		data := []byte("cookies are yummy")
		err = afs.WriteFile("/basedir/apps/read me.md", data, 0755)
		require.NoError(t, err)

		// when
		fsys, err := validation.NewInMemoryFS(logger, afs, "/basedir")

		// then
		require.NoError(t, err)

		assert.True(t, fsys.Exists("/basedir/apps"))
		assert.False(t, fsys.Exists("/basedir/apps/read me.md"))
	})
}
