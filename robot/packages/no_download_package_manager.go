package packages

import (
	"context"

	"go.viam.com/rdk/config"
)

type noDownloadManager struct {
	packagesDir string
}

var (
	_ Manager       = (*noDownloadManager)(nil)
	_ ManagerSyncer = (*noDownloadManager)(nil)
)

func NewNoDownloadManager() ManagerSyncer {
	return &noDownloadManager{}
}

func (m *noDownloadManager) PackagePath(name PackageName) (string, error) {
	return localNamedPath(name, m.packagesDir), nil
}

func (m *noDownloadManager) RefPath(refPath string) (string, error) {
	// TODO: update this and actually use it
	return refPath, nil
}

func (m *noDownloadManager) Close() error {
	return nil
}

func (m *noDownloadManager) Sync(ctx context.Context, packages []config.PackageConfig) error {
	return nil
}

func (m *noDownloadManager) Cleanup(ctx context.Context) error {
	return nil
}
