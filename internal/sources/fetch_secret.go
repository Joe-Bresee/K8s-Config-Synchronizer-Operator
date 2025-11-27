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

// FetchSecret fetches a Secret, writes its data keys to a temp dir (as files),
// and returns a deterministic revision SHA and the temp dir path.
// Secret values are treated as raw bytes; SHA is computed deterministically.
func fetchSecret(
	ctx context.Context,
	c client.Client,
	ref *configsv1alpha1.ObjectRef,
) (string, string, error) {
	logger := log.FromContext(ctx)

	if ref == nil {
		return "", "", fmt.Errorf("SecretRef cannot be nil")
	}

	var s corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &s); err != nil {
		return "", "", fmt.Errorf("failed to get Secret %s/%s: %w", ref.Namespace, ref.Name, err)
	}

	logger.Info("fetched Secret", "namespace", s.Namespace, "name", s.Name)

	// create temp dir for secret files
	tempDir, err := os.MkdirTemp("", "config-sync-secret-")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir for Secret data: %w", err)
	}

	// write each key as a file, restrictive permissions
	for filename, data := range s.Data {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, data, 0o600); err != nil {
			return "", "", fmt.Errorf("failed to write Secret data to file %s: %w", filePath, err)
		}
		logger.Info("wrote Secret data to file", "filePath", filePath)
	}

	// deterministic SHA: sort keys, include length, hex-encode values
	keys := make([]string, 0, len(s.Data))
	for k := range s.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, k := range keys {
		v := s.Data[k]
		buf.WriteString(k)
		buf.WriteByte('\n')
		buf.WriteString(fmt.Sprintf("%d", len(v)))
		buf.WriteByte('\n')
		buf.WriteString(hex.EncodeToString(v))
		buf.WriteByte('\n')
	}

	h := sha256.Sum256([]byte(buf.String()))
	sha := hex.EncodeToString(h[:])

	return sha, tempDir, nil
}
