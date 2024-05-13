package applications_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeready-toolchain/sandbox-argocd/pkg/applications"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestListApplications(t *testing.T) {

	baseDir := "/path/to/apps"

	t.Run("empty dir", func(t *testing.T) {
		// given
		afs, err := newAfs(baseDir)
		require.NoError(t, err)
		logger := log.New(os.Stdout)

		// when
		apps, appsets, err := applications.ListApplications(logger, afs, baseDir)

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
		apps, appsets, err := applications.ListApplications(logger, afs, baseDir)

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

func TestCreateApplication(t *testing.T) {

	ctx := context.TODO()
	logger := log.New(os.Stdout)
	s := scheme.Scheme
	err := argocdv1alpha1.AddToScheme(s)
	require.NoError(t, err)

	t.Run("create", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// given
			app := &argocdv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "openshift-gitops",
					Name:      "cookie",
				},
			}
			cl := fake.NewClientBuilder().
				WithScheme(s).
				Build()
			// when
			err := applications.CreateApplication(ctx, logger, cl, app)
			// then
			require.NoError(t, err)
		})

		t.Run("failure", func(t *testing.T) {
			// given
			app := &argocdv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "openshift-gitops",
					Name:      "cookie",
				},
			}
			cl := fake.NewClientBuilder().
				WithScheme(s).
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, cl runtimeclient.WithWatch, obj runtimeclient.Object, opts ...runtimeclient.CreateOption) error {
						if _, ok := obj.(*argocdv1alpha1.Application); ok {
							return fmt.Errorf("mock error!")
						}
						return cl.Create(ctx, obj, opts...)
					},
				}).
				Build()
			// when
			err := applications.CreateApplication(ctx, logger, cl, app)
			// then
			require.Error(t, err, "mock error!")
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// given
			app := &argocdv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "openshift-gitops",
					Name:      "cookie",
				},
			}
			existingApp := &argocdv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:       "openshift-gitops",
					Name:            "cookie",
					ResourceVersion: "1",
				},
			}
			cl := fake.NewClientBuilder().
				WithScheme(s).
				WithRuntimeObjects(existingApp).
				Build()
			// when
			err := applications.CreateApplication(ctx, logger, cl, app)
			// then
			require.NoError(t, err)
		})

		t.Run("failure", func(t *testing.T) {
			// given
			app := &argocdv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "openshift-gitops",
					Name:      "cookie",
				},
			}
			existingApp := &argocdv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:       "openshift-gitops",
					Name:            "cookie",
					ResourceVersion: "1",
				},
			}
			cl := fake.NewClientBuilder().
				WithScheme(s).
				WithRuntimeObjects(existingApp).
				WithInterceptorFuncs(interceptor.Funcs{
					Update: func(ctx context.Context, cl runtimeclient.WithWatch, obj runtimeclient.Object, opts ...runtimeclient.UpdateOption) error {
						if _, ok := obj.(*argocdv1alpha1.Application); ok {
							return fmt.Errorf("mock error!")
						}
						return cl.Update(ctx, obj, opts...)
					},
				}).
				Build()
			// when
			err := applications.CreateApplication(ctx, logger, cl, app)
			// then
			require.Error(t, err, "mock error!")
		})
	})
}

func TestCreateApplicationSet(t *testing.T) {

	ctx := context.TODO()
	logger := log.New(os.Stdout)
	s := scheme.Scheme
	err := argocdv1alpha1.AddToScheme(s)
	require.NoError(t, err)

	t.Run("create", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// given
			app := &argocdv1alpha1.ApplicationSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "openshift-gitops",
					Name:      "cookie",
				},
			}
			cl := fake.NewClientBuilder().
				WithScheme(s).
				Build()
			// when
			err := applications.CreateApplicationSet(ctx, logger, cl, app)
			// then
			require.NoError(t, err)
		})

		t.Run("failure", func(t *testing.T) {
			// given
			app := &argocdv1alpha1.ApplicationSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "openshift-gitops",
					Name:      "cookie",
				},
			}
			cl := fake.NewClientBuilder().
				WithScheme(s).
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, cl runtimeclient.WithWatch, obj runtimeclient.Object, opts ...runtimeclient.CreateOption) error {
						if _, ok := obj.(*argocdv1alpha1.ApplicationSet); ok {
							return fmt.Errorf("mock error!")
						}
						return cl.Create(ctx, obj, opts...)
					},
				}).
				Build()
			// when
			err := applications.CreateApplicationSet(ctx, logger, cl, app)
			// then
			require.Error(t, err, "mock error!")
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// given
			app := &argocdv1alpha1.ApplicationSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "openshift-gitops",
					Name:      "cookie",
				},
			}
			existingApp := &argocdv1alpha1.ApplicationSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:       "openshift-gitops",
					Name:            "cookie",
					ResourceVersion: "1",
				},
			}
			cl := fake.NewClientBuilder().
				WithScheme(s).
				WithRuntimeObjects(existingApp).
				Build()
			// when
			err := applications.CreateApplicationSet(ctx, logger, cl, app)
			// then
			require.NoError(t, err)
		})

		t.Run("failure", func(t *testing.T) {
			// given
			app := &argocdv1alpha1.ApplicationSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "openshift-gitops",
					Name:      "cookie",
				},
			}
			existingApp := &argocdv1alpha1.ApplicationSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:       "openshift-gitops",
					Name:            "cookie",
					ResourceVersion: "1",
				},
			}
			cl := fake.NewClientBuilder().
				WithScheme(s).
				WithRuntimeObjects(existingApp).
				WithInterceptorFuncs(interceptor.Funcs{
					Update: func(ctx context.Context, cl runtimeclient.WithWatch, obj runtimeclient.Object, opts ...runtimeclient.UpdateOption) error {
						if _, ok := obj.(*argocdv1alpha1.ApplicationSet); ok {
							return fmt.Errorf("mock error!")
						}
						return cl.Update(ctx, obj, opts...)
					},
				}).
				Build()
			// when
			err := applications.CreateApplicationSet(ctx, logger, cl, app)
			// then
			require.Error(t, err, "mock error!")
		})
	})
}
