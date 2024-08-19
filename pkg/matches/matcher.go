package matches

import "context"

type ContextKey string

// UpdateSnapshotKey defines a signal to the snapshot watcher to update the snapshot its checking.
var UpdateSnapshotKey = ContextKey("update-snapshot")

// Matcher that can assert information given a CRD and a payload configuration of the matcher.
type Matcher interface {
	Match(ctx context.Context, crdLocation string, payload []byte) error
}
