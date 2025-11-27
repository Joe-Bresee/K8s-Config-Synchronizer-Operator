package source

import (
	"context"
	"fmt"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FetchSource(configSync *configsv1alpha1.ConfigSync, ctx context.Context, c client.Client) (string, string, error) {
	sourceSet := 0
	if configSync.Spec.Source.Git != nil {
		sourceSet++
	}
	if configSync.Spec.Source.ConfigMapRef != nil {
		sourceSet++
	}
	if configSync.Spec.Source.SecretRef != nil {
		sourceSet++
	}
	if sourceSet != 1 {
		return "", "", fmt.Errorf("exactly one source must be specified in ConfigSync spec")
	}

	var revisionSHA string
	var sourcePath string

	if configSync.Spec.Source.Git != nil {
		var err error
		revisionSHA, sourcePath, err = cloneOrUpdate(
			ctx,
			c,
			configSync.Spec.Source.Git.RepoURL,
			configSync.Spec.Source.Git.Revision,
			configSync.Spec.Source.Git.Branch,
			configSync.Spec.Source.Git.AuthMethod,
			configSync.Spec.Source.Git.AuthSecretRef,
		)
		if err != nil {
			return "", "", fmt.Errorf("failed to clone or update git repository: %w", err)
		}

	}
	if configSync.Spec.Source.ConfigMapRef != nil {
		// Handle ConfigMap source logic here
		var err error
		revisionSHA, sourcePath, err = fetchConfigMap(
			ctx,
			c,
			configSync.Spec.Source.ConfigMapRef,
		)
		if err != nil {
			return "", "", fmt.Errorf("failed to fetch ConfigMap source: %w", err)
		}
	}
	if configSync.Spec.Source.SecretRef != nil {
		// Handle Secret source logic here
		var err error
		revisionSHA, sourcePath, err = fetchSecret(
			ctx,
			c,
			configSync.Spec.Source.SecretRef,
		)
		if err != nil {
			return "", "", fmt.Errorf("failed to fetch Secret source: %w", err)
		}
	}
	return revisionSHA, sourcePath, nil
}
