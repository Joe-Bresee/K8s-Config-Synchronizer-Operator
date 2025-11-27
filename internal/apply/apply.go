package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	"go.yaml.in/yaml/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ApplyTarget(ctx context.Context, c client.Client, sourcePath string, target configsv1alpha1.TargetRef) error {
	logger := log.FromContext(ctx)

	// 1. List all YAML files
	files, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to list source directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !(strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		filePath := filepath.Join(sourcePath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// 2. Handle multi-document YAML
		docs := strings.Split(string(data), "\n---")
		for _, doc := range docs {
			doc = strings.TrimSpace(doc)
			if doc == "" {
				continue
			}

			// 3. Decode into Unstructured object
			obj := &unstructured.Unstructured{}
			if err := yaml.Unmarshal([]byte(doc), obj); err != nil {
				return fmt.Errorf("failed to parse YAML in %s: %w", filePath, err)
			}

			// 4. Override namespace if object is namespaced
			if target.Namespace != "" {
				obj.SetNamespace(target.Namespace)
			}

			// 5. Dry-run validation: perform a server-side dry-run apply first to catch admission/validation errors
			applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("configsync")}
			dryRunOpts := append(applyOpts, client.DryRunAll)

			if err := c.Patch(ctx, obj, client.Apply, dryRunOpts...); err != nil {
				return fmt.Errorf("dry-run failed for %s from %s: %w",
					obj.GetKind(), filePath, err)
			}

			// 6. Actual apply
			if err := c.Patch(ctx, obj, client.Apply, applyOpts...); err != nil {
				return fmt.Errorf("failed to apply %s from %s: %w",
					obj.GetKind(), filePath, err)
			}

			logger.Info("Applied manifest",
				"kind", obj.GetKind(),
				"name", obj.GetName(),
				"namespace", obj.GetNamespace(),
				"file", filePath,
			)
		}
	}

	return nil
}

// No rendering (not Helm)

// No Kustomize

// No drift detection (SSA handles merge)

// No pruning of deleted files yet (add later if desired)
