import assert from "node:assert/strict";
import test from "node:test";

import { normalizeGraphStreamPayload } from "./viewer_protocol.mjs";

const TIMESTAMP = "2026-02-12T10:00:00Z";

function snapshot(id, dot = `digraph ${id} {}`) {
  return { id, timestamp: TIMESTAMP, dot };
}

function collection(id, snapshots) {
  return {
    id,
    timestamp: TIMESTAMP,
    snapshots,
  };
}

test("normalizeGraphStreamPayload filters malformed snapshot and collection data", () => {
  const normalized = normalizeGraphStreamPayload({
    workingSnapshots: [
      snapshot(1),
      { id: 2, timestamp: TIMESTAMP },
      null,
    ],
    pastCollections: [
      collection(10, [snapshot(7), { id: 8, timestamp: TIMESTAMP }]),
      { id: 11, timestamp: TIMESTAMP },
      null,
    ],
    latestWorkingId: "bad",
    latestPastCollectionId: 22,
  });

  assert.deepEqual(normalized.workingSnapshots, [snapshot(1)]);
  assert.deepEqual(normalized.pastCollections, [collection(10, [snapshot(7)])]);
  assert.equal(normalized.latestWorkingId, 0);
  assert.equal(normalized.latestPastCollectionId, 22);
});

test("normalizeGraphStreamPayload handles non-object input", () => {
  assert.deepEqual(normalizeGraphStreamPayload(null), {
    workingSnapshots: [],
    pastCollections: [],
  });
});
