package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

// DryRunEnabled controls whether ApplyTarget performs a server-side dry-run
var DryRunEnabled = true

func ApplyTarget(ctx context.Context, c client.Client, sourcePath string, target configsv1alpha1.TargetRef) error {
	logger := log.FromContext(ctx)

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

		docs := strings.SplitSeq(string(data), "\n---")
		for doc := range docs {
			doc = strings.TrimSpace(doc)
			if doc == "" {
				continue
			}

			obj := &unstructured.Unstructured{}
			jsonData, err := yaml.YAMLToJSON([]byte(doc))
			if err != nil {
				return fmt.Errorf("failed to convert YAML to JSON in %s: %w", filePath, err)
			}
			if err := obj.UnmarshalJSON(jsonData); err != nil {
				return fmt.Errorf("failed to unmarshal object in %s: %w", filePath, err)
			}

			if target.Namespace != "" {
				obj.SetNamespace(target.Namespace)
			}

			// helper function to ensure metadata doesn't exist in the object and other cleanup to ensure k apply accepts obj
			cleanObjectForApply(obj)

			applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("configsync")}
			dryRunOpts := append(applyOpts, client.DryRunAll)

			if DryRunEnabled {
				logger.Info("Performing dry-run apply")
				if err := c.Patch(ctx, obj, client.Apply, dryRunOpts...); err != nil {
					return fmt.Errorf("dry-run failed for %s from %s: %w",
						obj.GetKind(), filePath, err)
				}
			}

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

// cleanObjectForApply removes metadata fields that must not be sent to the API server
func cleanObjectForApply(obj *unstructured.Unstructured) {
	if obj == nil {
		return
	}
	content := obj.UnstructuredContent()

	// Remove top-level status if present
	delete(content, "status")

	// Recursively remove server-populated fields
	removeManagedFieldsRecursive(content)

	// Write back content
	obj.SetUnstructuredContent(content)
}

// removeManagedFieldsRecursive removes managedFields and other server-populated metadata recursively
func removeManagedFieldsRecursive(v interface{}) {
	switch t := v.(type) {
	case map[string]interface{}:
		// Remove metadata.managedFields and common server-populated fields. maybe there's a better way to do all of this?
		if metaI, ok := t["metadata"]; ok {
			if meta, ok := metaI.(map[string]interface{}); ok {
				delete(meta, "managedFields")
				delete(meta, "resourceVersion")
				delete(meta, "uid")
				delete(meta, "creationTimestamp")
				delete(meta, "generation")
				delete(meta, "selfLink")
				t["metadata"] = meta
			}
		}
		// Remove any top-level managedFields key
		delete(t, "managedFields")
		for _, v := range t {
			removeManagedFieldsRecursive(v)
		}
	case []interface{}:
		for _, e := range t {
			removeManagedFieldsRecursive(e)
		}
	}
}
