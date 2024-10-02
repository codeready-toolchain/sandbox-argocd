package validation_test

import (
	"os"
	"testing"

	"github.com/codeready-toolchain/sandbox-argocd/pkg/validation"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestCheckComponents(t *testing.T) {

	t.Run("success", func(t *testing.T) {

		t.Run("no component", func(t *testing.T) {
			// given
			logger := log.New(os.Stdout)
			afs := afero.Afero{
				Fs: afero.NewMemMapFs(),
			}
			err := afs.Mkdir("/path/to/components", os.ModeDir)
			require.NoError(t, err)

			// when
			err = validation.CheckComponents(logger, afs, "/path/to", "components")

			// then
			require.NoError(t, err)
		})

		t.Run("empty component", func(t *testing.T) {
			// given
			logger := log.New(os.Stdout)
			afs := afero.Afero{
				Fs: afero.NewMemMapFs(),
			}
			err := afs.Mkdir("/path/to/components", os.ModeDir)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/kustomization.yaml", `kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1`)
			require.NoError(t, err)

			// when
			err = validation.CheckComponents(logger, afs, "/path/to", "components")

			// then
			require.Error(t, err, "invalid resources at /path/to/components: kustomization.yaml is empty")
		})

		t.Run("component with secretGenerator", func(t *testing.T) {
			// given
			logger := log.New(os.Stdout)
			afs := afero.Afero{
				Fs: afero.NewMemMapFs(),
			}
			err := afs.Mkdir("/path/to/components", os.ModeDir)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/kustomization.yaml", `kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

secretGenerator:
  - name: mysecret1
    files:
      - secret1.yaml
  - name: mysecret2
    files:
      - secret2=secret2.yaml`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/secret1.yaml", `apiVersion: v1
kind: Secret
metadata:
  namespace: test
  name: secret
data:
  cookie: yummy`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/secret2.yaml", `apiVersion: v1
kind: Secret
metadata:
  namespace: test
  name: secret
data:
  pasta: yummy`)
			require.NoError(t, err)
			// when
			err = validation.CheckComponents(logger, afs, "/path/to", "components")

			// then
			require.NoError(t, err)
		})

		t.Run("component with configmapGenerator", func(t *testing.T) {
			// given
			logger := log.New(os.Stdout)
			afs := afero.Afero{
				Fs: afero.NewMemMapFs(),
			}
			err := afs.Mkdir("/path/to/components", os.ModeDir)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/kustomization.yaml", `kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

configMapGenerator:
  - name: myconfig1
    files:
      - configmap1.yaml
  - name: myconfig2
    files:
      - cm=configmap2.yaml`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/configmap1.yaml", `apiVersion: v1
kind: ConfigMap
metadata:
  namespace: test
  name: cm1
data:
  cookie: yummy`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/configmap2.yaml", `apiVersion: v1
kind: ConfigMap
metadata:
  namespace: test
  name: cm2
data:
  cookie: yummy`)
			require.NoError(t, err)
			// when
			err = validation.CheckComponents(logger, afs, "/path/to", "components")

			// then
			require.NoError(t, err)
		})

		t.Run("component with patchesStrategicMerge", func(t *testing.T) {
			// given
			logger := log.New(os.Stdout)
			afs := afero.Afero{
				Fs: afero.NewMemMapFs(),
			}
			err := afs.Mkdir("/path/to/components", os.ModeDir)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/kustomization.yaml", `kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- deployment.yaml

patchesStrategicMerge:
  - patch.yaml`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/deployment.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: test
  name: test
spec:
  replicas: 1
`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/patch.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: test
  name: test
spec:
  replicas: 2
`)
			require.NoError(t, err)

			// when
			err = validation.CheckComponents(logger, afs, "/path/to", "components")

			// then
			require.NoError(t, err)
		})

		t.Run("component with patches", func(t *testing.T) {
			// given
			logger := log.New(os.Stdout)
			afs := afero.Afero{
				Fs: afero.NewMemMapFs(),
			}
			err := afs.Mkdir("/path/to/components", os.ModeDir)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/kustomization.yaml", `kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- deployment.yaml

patches:
  - path: patch.yaml`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/deployment.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: test
  name: test
spec:
  replicas: 1
`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/patch.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: test
  name: test
spec:
  replicas: 2
`)
			require.NoError(t, err)

			// when
			err = validation.CheckComponents(logger, afs, "/path/to", "components")

			// then
			require.NoError(t, err)
		})

		t.Run("component with transformers", func(t *testing.T) {
			// given
			logger := log.New(os.Stdout)
			afs := afero.Afero{
				Fs: afero.NewMemMapFs(),
			}
			err := afs.Mkdir("/path/to/components", os.ModeDir)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/kustomization.yaml", `kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- deployment.yaml

transformers:
  - namespace.yaml`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/deployment.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: test
  name: test
spec:
  replicas: 1
`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/namespace.yaml", `apiVersion: builtin
kind: NamespaceTransformer
metadata:
  name: set-namespace
  namespace: foo
setRoleBindingSubjects: defaultOnly
`)
			require.NoError(t, err)

			// when
			err = validation.CheckComponents(logger, afs, "/path/to", "components")

			// then
			require.NoError(t, err)
		})
	})

	t.Run("warning", func(t *testing.T) {

		t.Run("component with unused resource", func(t *testing.T) {
			// given
			logger := log.New(os.Stdout)
			afs := afero.Afero{
				Fs: afero.NewMemMapFs(),
			}
			err := afs.Mkdir("/path/to/components", os.ModeDir)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/kustomization.yaml", `kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
  - configmap1.yaml`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/configmap1.yaml", `apiVersion: v1
kind: ConfigMap
metadata:
  namespace: test
  name: config1
data:
  cookie: yummy`)
			require.NoError(t, err)
			err = addFile(afs, "/path/to/components/configmap2.yaml", `apiVersion: v1
kind: ConfigMap
metadata:
  namespace: test
  name: config2
data:
  cookie: yummy`)
			require.NoError(t, err)

			// when
			err = validation.CheckComponents(logger, afs, "/path/to", "components")

			// then
			require.EqualError(t, err, "resource is not referenced in components/kustomization.yaml: configmap2.yaml")
		})
	})
}
