// Package snapshot manages metadata about snapshots stored in repository.
package snapshot

import (
	"context"
	"fmt"

	"github.com/kopia/kopia/internal/kopialogging"
	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/manifest"
)

var log = kopialogging.Logger("kopia/snapshot")

// ListSources lists all snapshot sources in a given repository.
func ListSources(ctx context.Context, rep *repo.Repository) ([]SourceInfo, error) {
	items, err := rep.Manifests.Find(ctx, map[string]string{
		"type": "snapshot",
	})
	if err != nil {
		return nil, fmt.Errorf("unable to find manifest entries: %v", err)
	}

	uniq := map[SourceInfo]bool{}
	for _, it := range items {
		uniq[sourceInfoFromLabels(it.Labels)] = true
	}

	var infos []SourceInfo
	for k := range uniq {
		infos = append(infos, k)
	}

	return infos, nil
}

func sourceInfoFromLabels(labels map[string]string) SourceInfo {
	return SourceInfo{Host: labels["hostname"], UserName: labels["username"], Path: labels["path"]}
}

func sourceInfoToLabels(si SourceInfo) map[string]string {
	return map[string]string{
		"type":     "snapshot",
		"hostname": si.Host,
		"username": si.UserName,
		"path":     si.Path,
	}
}

// ListSnapshots lists all snapshots for a given source.
func ListSnapshots(ctx context.Context, rep *repo.Repository, si SourceInfo) ([]*Manifest, error) {
	entries, err := rep.Manifests.Find(ctx, sourceInfoToLabels(si))
	if err != nil {
		return nil, fmt.Errorf("unable to find manifest entries: %v", err)
	}
	return LoadSnapshots(ctx, rep, manifest.EntryIDs(entries))
}

// LoadSnapshot loads and parses a snapshot with a given ID.
func LoadSnapshot(ctx context.Context, rep *repo.Repository, manifestID string) (*Manifest, error) {
	sm := &Manifest{}
	if err := rep.Manifests.Get(ctx, manifestID, sm); err != nil {
		return nil, fmt.Errorf("unable to find manifest entries: %v", err)
	}

	sm.ID = manifestID

	return sm, nil
}

// SaveSnapshot persists given snapshot manifest and returns manifest ID.
func SaveSnapshot(ctx context.Context, rep *repo.Repository, manifest *Manifest) (string, error) {
	return rep.Manifests.Put(ctx, sourceInfoToLabels(manifest.Source), manifest)
}

// LoadSnapshots efficiently loads and parses a given list of snapshot IDs.
func LoadSnapshots(ctx context.Context, rep *repo.Repository, names []string) ([]*Manifest, error) {
	result := make([]*Manifest, len(names))
	sem := make(chan bool, 50)

	for i, n := range names {
		sem <- true
		go func(i int, n string) {
			defer func() { <-sem }()

			m, err := LoadSnapshot(ctx, rep, n)
			if err != nil {
				log.Warningf("unable to parse snapshot manifest %v: %v", n, err)
				return
			}
			result[i] = m
		}(i, n)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	close(sem)

	successful := result[:0]
	for _, m := range result {
		if m != nil {
			successful = append(successful, m)
		}
	}

	return successful, nil
}

// ListSnapshotManifests returns the list of snapshot manifests for a given source or all sources if nil.
func ListSnapshotManifests(ctx context.Context, rep *repo.Repository, src *SourceInfo) ([]string, error) {
	labels := map[string]string{
		"type": "snapshot",
	}

	if src != nil {
		labels = sourceInfoToLabels(*src)
	}

	entries, err := rep.Manifests.Find(ctx, labels)
	if err != nil {
		return nil, fmt.Errorf("unable to find manifest entries: %v", err)
	}

	return manifest.EntryIDs(entries), nil
}
