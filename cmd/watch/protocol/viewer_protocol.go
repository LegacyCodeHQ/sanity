package protocol

import "time"

const (
	RouteIndex  = "/"
	RouteEvents = "/events"
)

const SSEEventGraph = "graph"

// GraphSnapshot is the atom in the watch protocol timeline.
type GraphSnapshot struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	DOT       string    `json:"dot"`
}

// GraphStreamPayload is the wire payload for SSE "graph" events.
type GraphStreamPayload struct {
	WorkingSnapshots       []GraphSnapshot      `json:"workingSnapshots"`
	PastCollections        []SnapshotCollection `json:"pastCollections"`
	LatestWorkingID        int64                `json:"latestWorkingId"`
	LatestPastCollectionID int64                `json:"latestPastCollectionId"`
}

// SnapshotCollection represents an archived batch of working snapshots.
type SnapshotCollection struct {
	ID        int64           `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	Snapshots []GraphSnapshot `json:"snapshots"`
}
