package factorymsg

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Namespace is a validated project namespace for one hook invocation. Its
// private fields prevent callers from substituting a directory or project key.
// Resolve it anew on the next invocation; repository identity is not cached.
type Namespace struct {
	root string
	dir  string
	key  string
}

// ResolveNamespace validates repository identity within the caller's budget.
// @MX:ANCHOR: [AUTO] one invocation's canonical identity shared by registry, broker, and cache paths
// @MX:REASON: registration and both broker opens require the same validated namespace without repeated Git discovery
func ResolveNamespace(ctx context.Context, root string) (Namespace, error) {
	if err := ctx.Err(); err != nil {
		return Namespace{}, err
	}
	dir, err := homestate.FactoryDirContext(ctx, root)
	if err != nil {
		return Namespace{}, err
	}
	if err := ctx.Err(); err != nil {
		return Namespace{}, err
	}
	return Namespace{root: root, dir: dir, key: filepath.Base(filepath.Dir(dir))}, nil
}

// FactoryDir returns the namespace already resolved by ResolveNamespace.
func (n Namespace) FactoryDir() string { return n.dir }

// ProjectKey returns the key derived by the canonical homestate resolver.
func (n Namespace) ProjectKey() string { return n.key }

// FactoryDBPath returns the registry beside this namespace's message brokers.
func (n Namespace) FactoryDBPath() string {
	if n.dir == "" {
		return ""
	}
	return filepath.Join(n.dir, "factory.db")
}

// OpenWithContext opens a broker without resolving this invocation's identity again.
func (n Namespace) OpenWithContext(parent context.Context, runID string) (*Store, error) {
	if err := parent.Err(); err != nil {
		return nil, err
	}
	if n.dir == "" || n.key == "" {
		return nil, errors.New("unresolved factory namespace")
	}
	if !safeID.MatchString(runID) {
		return nil, errors.New("invalid factory run id")
	}
	ctx, cancel := context.WithTimeout(parent, defaultBrokerOpenDeadline)
	defer cancel()
	return openResolvedWithContext(ctx, n, runID, nil)
}
