package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/kopia/kopia/repo/block"
	"github.com/kopia/kopia/repo/manifest"
	"github.com/kopia/kopia/repo/object"
	"github.com/kopia/kopia/repo/storage"
)

// Repository represents storage where both content-addressable and user-addressable data is kept.
type Repository struct {
	Blocks    *block.Manager
	Objects   *object.Manager
	Storage   storage.Storage
	Manifests *manifest.Manager
	UniqueID  []byte

	ConfigFile     string
	CacheDirectory string
}

// Close closes the repository and releases all resources.
func (r *Repository) Close(ctx context.Context) error {
	if err := r.Manifests.Flush(ctx); err != nil {
		return err
	}
	if err := r.Objects.Close(ctx); err != nil {
		return err
	}
	if err := r.Blocks.Flush(ctx); err != nil {
		return err
	}
	if err := r.Storage.Close(ctx); err != nil {
		return err
	}
	return nil
}

// Flush waits for all in-flight writes to complete.
func (r *Repository) Flush(ctx context.Context) error {
	if err := r.Manifests.Flush(ctx); err != nil {
		return err
	}
	if err := r.Objects.Flush(ctx); err != nil {
		return err
	}

	return r.Blocks.Flush(ctx)
}

// Refresh periodically makes external changes visible to repository.
func (r *Repository) Refresh(ctx context.Context) error {
	updated, err := r.Blocks.Refresh(ctx)
	if err != nil {
		return fmt.Errorf("error refreshing block index: %v", err)
	}

	if !updated {
		return nil
	}

	log.Debugf("block index refreshed")

	if err := r.Manifests.Refresh(ctx); err != nil {
		return fmt.Errorf("error reloading manifests: %v", err)
	}

	log.Debugf("manifests refreshed")

	return nil
}

// RefreshPeriodically periodically refreshes the repository to reflect the changes made by other hosts.
func (r *Repository) RefreshPeriodically(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(interval):
			if err := r.Refresh(ctx); err != nil {
				log.Warningf("error refreshing repository: %v", err)
			}
		}
	}
}
