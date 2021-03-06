package policy

import (
	"context"
	"strings"

	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/snapshot"
)

// GetExpiredSnapshots computes the set of snapshot manifests that are not retained according to the policy.
func GetExpiredSnapshots(ctx context.Context, rep *repo.Repository, snapshots []*snapshot.Manifest) ([]*snapshot.Manifest, error) {
	var toDelete []*snapshot.Manifest
	for _, snapshotGroup := range snapshot.GroupBySource(snapshots) {
		td, err := getExpiredSnapshotsForSource(ctx, rep, snapshotGroup)
		if err != nil {
			return nil, err
		}
		toDelete = append(toDelete, td...)
	}
	return toDelete, nil
}

func getExpiredSnapshotsForSource(ctx context.Context, rep *repo.Repository, snapshots []*snapshot.Manifest) ([]*snapshot.Manifest, error) {
	src := snapshots[0].Source
	pol, _, err := GetEffectivePolicy(ctx, rep, src)
	if err != nil {
		return nil, err
	}

	pol.RetentionPolicy.ComputeRetentionReasons(snapshots)

	var toDelete []*snapshot.Manifest
	for _, s := range snapshots {
		if len(s.RetentionReasons) == 0 {
			log.Debugf("  deleting %v", s.StartTime)
			toDelete = append(toDelete, s)
		} else {
			log.Debugf("  keeping %v reasons: [%v]", s.StartTime, strings.Join(s.RetentionReasons, ","))
		}
	}
	return toDelete, nil
}
