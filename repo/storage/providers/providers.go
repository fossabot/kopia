// Package providers registers all storage providers that are included as part of Kopia.
package providers

import (
	// Register well-known blob storage providers
	_ "github.com/kopia/kopia/repo/storage/filesystem"
	_ "github.com/kopia/kopia/repo/storage/gcs"
)
