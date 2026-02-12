import { Graphviz } from "https://cdn.jsdelivr.net/npm/@hpcc-js/wasm-graphviz@1.6.1/dist/index.js";

const graphviz = await Graphviz.load();
const container = document.getElementById("graph-container");
const statusEl = document.getElementById("status");
const statusText = document.getElementById("status-text");
const sliderEl = document.getElementById("timeline-slider");
const liveBtnEl = document.getElementById("timeline-live");
const modeEl = document.getElementById("timeline-mode");
const metaEl = document.getElementById("timeline-meta");

let snapshots = [];
let manualMode = false;

function renderGraph(dot) {
  try {
    const svg = graphviz.layout(dot, "svg", "dot");
    container.innerHTML = svg;
    statusText.textContent = "Connected";
    statusEl.classList.remove("disconnected");
  } catch (err) {
    console.error("Graphviz render error:", err);
    statusText.textContent = "Render error";
  }
}

function renderWaitingState() {
  container.innerHTML = '<p id="placeholder">No uncommitted changes. Waiting for file changes...</p>';
}

function formatSnapshotMeta(snapshot, index) {
  const time = new Date(snapshot.timestamp);
  return `#${index + 1} ${time.toLocaleTimeString()}`;
}

function syncTimelineUI() {
  const total = snapshots.length;
  sliderEl.disabled = total <= 1;
  sliderEl.max = total > 0 ? String(total - 1) : "0";
  if (total === 0) {
    sliderEl.value = "0";
  }

  if (!manualMode && total > 0) {
    sliderEl.value = String(total - 1);
  }

  const selected = Number(sliderEl.value || "0");
  const snapshot = snapshots[selected];
  const modeText = manualMode ? "Snapshot" : "Latest";
  modeEl.textContent = modeText;
  liveBtnEl.disabled = !manualMode || total === 0;

  if (snapshot) {
    metaEl.textContent = `${total} snapshots | ${formatSnapshotMeta(snapshot, selected)}`;
  } else {
    metaEl.textContent = "0 snapshots";
  }
}

function renderSelectedSnapshot() {
  if (snapshots.length === 0) {
    syncTimelineUI();
    renderWaitingState();
    return;
  }
  const idx = manualMode ? Number(sliderEl.value || "0") : snapshots.length - 1;
  const snapshot = snapshots[idx];
  if (!snapshot) {
    return;
  }
  renderGraph(snapshot.dot);
  syncTimelineUI();
}

function mergePayload(payload) {
  snapshots = payload.snapshots || [];
  if (snapshots.length === 0) {
    manualMode = false;
    renderSelectedSnapshot();
    return;
  }

  if (!manualMode) {
    renderSelectedSnapshot();
    return;
  }

  const maxIdx = Math.max(snapshots.length - 1, 0);
  sliderEl.value = String(Math.min(Number(sliderEl.value || "0"), maxIdx));
  renderSelectedSnapshot();
}

sliderEl.addEventListener("input", function() {
  if (snapshots.length === 0) {
    return;
  }
  manualMode = true;
  renderSelectedSnapshot();
});

liveBtnEl.addEventListener("click", function() {
  manualMode = false;
  renderSelectedSnapshot();
});

function connectSSE() {
  const source = new EventSource("/events");

  source.addEventListener("graph", function(event) {
    try {
      const payload = JSON.parse(event.data);
      mergePayload(payload);
    } catch (err) {
      console.error("Invalid graph payload:", err);
      statusText.textContent = "Payload error";
    }
  });

  source.addEventListener("open", function() {
    statusText.textContent = "Connected";
    statusEl.classList.remove("disconnected");
  });

  source.addEventListener("error", function() {
    statusText.textContent = "Reconnecting...";
    statusEl.classList.add("disconnected");
  });
}

connectSSE();
