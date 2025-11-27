package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gittransport "github.com/go-git/go-git/v5/plumbing/transport"
	httpAuth "github.com/go-git/go-git/v5/plumbing/transport/http"
	sshAuth "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	configsv1alpha1 "github.com/joe-bresee/config-synchronizer-operator/api/v1alpha1"
)

func cloneOrUpdate(
	ctx context.Context,
	c client.Client,
	repoURL, revision, branch, authMethod string,
	authSecretRef *configsv1alpha1.ObjectRef,
) (string, string, error) {

	logger := log.FromContext(ctx)

	cacheBase := filepath.Join(os.TempDir(), "config-sync-cache")
	repoName := sanitizeRepoURL(repoURL)
	cachePath := filepath.Join(cacheBase, repoName)

	logger.Info("preparing repository cache", "path", cachePath)

	// Ensure base dir exists
	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		return "", "", fmt.Errorf("failed to create cache path: %w", err)
	}

	// Build auth if needed
	authMethodObj, err := buildAuth(ctx, c, authMethod, authSecretRef)
	if err != nil {
		return "", "", err
	}

	// Determine clone vs open
	var repo *git.Repository

	gitDir := filepath.Join(cachePath, ".git")
	_, statErr := os.Stat(gitDir)

	if os.IsNotExist(statErr) {
		// Clone fresh
		logger.Info("cloning repository", "url", repoURL, "branch", branch)

		cloneOpts := &git.CloneOptions{
			URL:  repoURL,
			Tags: git.AllTags,
		}
		if authMethodObj != nil {
			cloneOpts.Auth = authMethodObj
		}
		if branch != "" {
			cloneOpts.SingleBranch = true
			cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(branch)
		}

		repo, err = git.PlainClone(cachePath, false, cloneOpts)
		if err != nil {
			return "", "", fmt.Errorf("clone failed: %w", err)
		}

	} else {
		// Repo exists → open & fetch
		logger.Info("opening cached repository", "path", cachePath)

		repo, err = git.PlainOpen(cachePath)
		if err != nil {
			logger.Error(err, "cached repo appears corrupted; removing")
			_ = os.RemoveAll(cachePath)
			return cloneOrUpdate(ctx, c, repoURL, revision, branch, authMethod, authSecretRef)
		}

		logger.Info("fetching latest updates from origin")

		fetchOpts := &git.FetchOptions{
			RemoteName: "origin",
			Tags:       git.AllTags,
			Force:      true,
		}

		if authMethodObj != nil {
			fetchOpts.Auth = authMethodObj
		}

		err = repo.Fetch(fetchOpts)
		if err != nil && err != git.NoErrAlreadyUpToDate {
			logger.Error(err, "fetch failed; repository may be corrupted")
			_ = os.RemoveAll(cachePath)
			return cloneOrUpdate(ctx, c, repoURL, revision, branch, authMethod, authSecretRef)
		}
	}

	w, err := repo.Worktree()
	if err != nil {
		return "", "", fmt.Errorf("failed to get worktree: %w", err)
	}

	if revision != "" {
		logger.Info("checking out revision", "revision", revision)
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(revision),
		})
		if err != nil {
			return "", "", fmt.Errorf("checkout revision failed: %w", err)
		}
	} else if branch != "" {
		logger.Info("checking out branch", "branch", branch)
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
		})
		if err != nil {
			return "", "", fmt.Errorf("checkout branch failed: %w", err)
		}
	}

	head, err := repo.Head()
	if err != nil {
		return "", "", fmt.Errorf("failed to read HEAD: %w", err)
	}

	logger.Info("repository synced",
		"repo", repoURL,
		"commit", head.Hash().String(),
	)

	return head.Hash().String(), cachePath, nil
}

func sanitizeRepoURL(url string) string {
	u := strings.ReplaceAll(url, "://", "_")
	u = strings.ReplaceAll(u, "/", "_")
	u = strings.ReplaceAll(u, "@", "_")
	u = strings.ReplaceAll(u, ":", "_")
	return u
}

func firstSecretKey(data map[string][]byte, keys []string) []byte {
	for _, k := range keys {
		if v, ok := data[k]; ok && len(v) > 0 {
			return v
		}
	}
	return nil
}

func firstSecretString(data map[string][]byte, keys []string) string {
	b := firstSecretKey(data, keys)
	if b == nil {
		return ""
	}
	return string(b)
}

func buildAuth(
	ctx context.Context,
	c client.Client,
	authMethod string,
	authSecretRef *configsv1alpha1.ObjectRef,
) (gittransport.AuthMethod, error) {

	logger := log.FromContext(ctx)

	if strings.EqualFold(authMethod, "none") || authMethod == "" {
		return nil, nil
	}

	if authSecretRef == nil {
		return nil, fmt.Errorf("authSecretRef required for authMethod=%s", authMethod)
	}

	var secret corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: authSecretRef.Namespace,
		Name:      authSecretRef.Name,
	}, &secret); err != nil {
		return nil, fmt.Errorf("failed to read Secret %s/%s: %w",
			authSecretRef.Namespace, authSecretRef.Name, err)
	}

	switch strings.ToLower(authMethod) {
	case "https":
		username := firstSecretString(secret.Data, []string{"username", "user"})
		password := firstSecretString(secret.Data, []string{"password", "pass", "token"})

		if username == "" || password == "" {
			return nil, fmt.Errorf("secret %s/%s missing username/password for https auth",
				authSecretRef.Namespace, authSecretRef.Name)
		}

		logger.Info("using HTTPS basic auth: found username and password", "namespace", authSecretRef.Namespace, "name", authSecretRef.Name)
		return &httpAuth.BasicAuth{
			Username: username,
			Password: password,
		}, nil

	case "ssh":
		key := firstSecretKey(secret.Data, []string{"sshKey", "id_rsa", "ssh-privatekey", "private_key"})
		if key == nil {
			return nil, fmt.Errorf("secret %s/%s missing private key for ssh auth",
				authSecretRef.Namespace, authSecretRef.Name)
		}

		logger.Info("using SSH key auth: found private key", "namespace", authSecretRef.Namespace, "name", authSecretRef.Name)

		signer, err := sshAuth.NewPublicKeys("git", key, "")
		if err != nil {
			return nil, fmt.Errorf("failed to load ssh private key: %w", err)
		}

		// ADD SUPPORT FOR KNOWN HOSTS
		signer.HostKeyCallback = ssh.InsecureIgnoreHostKey()

		return signer, nil

	default:
		return nil, fmt.Errorf("unsupported auth method, or something went wrong: %s", authMethod)
	}
}
