function normalizeSnapshot(snapshot) {
  if (!snapshot || typeof snapshot !== "object") {
    return null;
  }
  if (typeof snapshot.dot !== "string") {
    return null;
  }

  return {
    id: Number.isFinite(snapshot.id) ? snapshot.id : 0,
    timestamp: typeof snapshot.timestamp === "string" ? snapshot.timestamp : new Date(0).toISOString(),
    dot: snapshot.dot,
  };
}

function normalizeCollection(collection) {
  if (!collection || typeof collection !== "object" || !Array.isArray(collection.snapshots)) {
    return null;
  }

  return {
    id: Number.isFinite(collection.id) ? collection.id : 0,
    timestamp: typeof collection.timestamp === "string" ? collection.timestamp : new Date(0).toISOString(),
    snapshots: collection.snapshots
      .map(normalizeSnapshot)
      .filter((snapshot) => snapshot !== null),
  };
}

/**
 * Normalizes untrusted SSE JSON payloads from the watch server.
 * Keeps only fields needed by the viewer state machine.
 */
export function normalizeGraphStreamPayload(payload) {
  if (!payload || typeof payload !== "object") {
    return {
      workingSnapshots: [],
      pastCollections: [],
    };
  }

  return {
    workingSnapshots: Array.isArray(payload.workingSnapshots)
      ? payload.workingSnapshots.map(normalizeSnapshot).filter((snapshot) => snapshot !== null)
      : [],
    pastCollections: Array.isArray(payload.pastCollections)
      ? payload.pastCollections.map(normalizeCollection).filter((collection) => collection !== null)
      : [],
    latestWorkingId: Number.isFinite(payload.latestWorkingId) ? payload.latestWorkingId : 0,
    latestPastCollectionId: Number.isFinite(payload.latestPastCollectionId) ? payload.latestPastCollectionId : 0,
  };
}
