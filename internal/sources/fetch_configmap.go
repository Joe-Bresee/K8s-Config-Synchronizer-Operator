package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// FetchConfigMap fetches a ConfigMap, writes its keys to a temp dir, and returns a deterministic revision SHA and the path.
func fetchConfigMap(
	ctx context.Context,
	c client.Client,
	ref *configsv1alpha1.ObjectRef,
) (string, string, error) {
	logger := log.FromContext(ctx)

	if ref == nil {
		return "", "", fmt.Errorf("ConfigMapRef cannot be nil")
	}

	// fetch the ConfigMap
	var cm corev1.ConfigMap
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}, &cm); err != nil {
		return "", "", fmt.Errorf("failed to get ConfigMap %s/%s: %w",
			ref.Namespace, ref.Name, err)
	}

	logger.Info("fetched ConfigMap", "namespace", cm.Namespace, "name", cm.Name)

	// create a temp dir to store the files
	tempDir, err := os.MkdirTemp("", "config-sync-configmap-")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir for ConfigMap data: %w", err)
	}

	// write each key to a file
	for filename, data := range cm.Data {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(data), 0o644); err != nil {
			return "", "", fmt.Errorf("failed to write ConfigMap data to file %s: %w", filePath, err)
		}
		logger.Info("wrote ConfigMap data to file", "filePath", filePath)
	}

	// generate revisionSHA based on sorted keys and their content
	keys := make([]string, 0, len(cm.Data))
	for k := range cm.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('\n')
		b.WriteString(cm.Data[k])
		b.WriteByte('\n')
	}

	payload := []byte(b.String())
	h := sha256.Sum256(payload)
	sha := hex.EncodeToString(h[:])

	return sha, tempDir, nil
}
