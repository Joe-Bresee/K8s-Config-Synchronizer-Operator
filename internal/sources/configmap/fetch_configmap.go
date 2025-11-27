package sources

import (
	"context"
	"fmt"
	"os"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// functionality for fetching configmap data and storing in tmp dir for syncing

func FetchConfigMap(
	ctx context.Context,
	c client.Client,
	ref *configsv1alpha1.ObjectRef,
) (string, error) {
	logger := log.FromContext(ctx)

	if ref == nil {
		return "", fmt.Errorf("ConfigMapRef cannot be nil")
	}

	// fetch tha cm
	var cm corev1.ConfigMap
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}, &cm); err != nil {
		return "", fmt.Errorf("failed to get ConfigMap %s/%s: %w",
			ref.Namespace, ref.Name, err)
	}

	logger.Info("fetched ConfigMap", "namespace", cm.Namespace, "name", cm.Name)

	// create a temp dir to store the files
	tempDir, err := os.MkdirTemp("", "config-sync-configmap-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir for ConfigMap data: %w", err)
	}

	if err := os.RemoveAll(tempDir); err != nil {
		return "", fmt.Errorf("failed to clean temp dir before writing ConfigMap data: %w", err)
	}

	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to recreate temp dir for ConfigMap data: %w", err)
	}

	// write each key to a file
	for filename, data := range cm.Data {
		filePath := fmt.Sprintf("%s/%s", tempDir, filename)
		if err := os.WriteFile(filePath, []byte(data), 0o644); err != nil {
			return "", fmt.Errorf("failed to write ConfigMap data to file %s: %w", filePath, err)
		}
		logger.Info("wrote ConfigMap data to file", "filePath", filePath)
	}

	return tempDir, nil
}
