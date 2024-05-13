package applications

import (
	"context"
	fs "io/fs"
	"path/filepath"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func ListApplications(logger *log.Logger, afs afero.Afero, baseDir string) ([]*argocdv1alpha1.Application, []*argocdv1alpha1.ApplicationSet, error) {
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

func CreateApplication(ctx context.Context, logger *log.Logger, cl runtimeclient.Client, app *argocdv1alpha1.Application) error {
	existingApp := &argocdv1alpha1.Application{}
	if err := cl.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, existingApp); err == nil {
		// app already exist, let's update it instead
		app.ResourceVersion = existingApp.ResourceVersion
		if err := cl.Update(ctx, app, &runtimeclient.UpdateOptions{}); err != nil {
			return err
		}
		logger.Infof("successfully updated the '%s' Argo CD Application", app.Name)
		return nil
	}
	if err := cl.Create(ctx, app, &runtimeclient.CreateOptions{}); err != nil {
		return err
	}
	logger.Infof("successfully created the '%s' Argo CD Application", app.Name)
	return nil
}

func CreateApplicationSet(ctx context.Context, logger *log.Logger, cl runtimeclient.Client, appset *argocdv1alpha1.ApplicationSet) error {
	existingAppset := &argocdv1alpha1.ApplicationSet{}
	if err := cl.Get(ctx, types.NamespacedName{
		Namespace: appset.Namespace,
		Name:      appset.Name,
	}, existingAppset); err == nil {
		// app already exist, let's update it instead
		appset.ResourceVersion = existingAppset.ResourceVersion
		if err := cl.Update(ctx, appset, &runtimeclient.UpdateOptions{}); err != nil {
			return err
		}
		logger.Infof("successfully updated the '%s' Argo CD ApplicationSet", appset.Name)
		return nil
	}
	if err := cl.Create(ctx, appset, &runtimeclient.CreateOptions{}); err != nil {
		return err
	}
	logger.Infof("successfully created the '%s' Argo CD ApplicationSet", appset.Name)
	return nil
}
