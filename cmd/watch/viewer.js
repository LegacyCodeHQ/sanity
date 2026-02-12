import { Graphviz } from "https://cdn.jsdelivr.net/npm/@hpcc-js/wasm-graphviz@1.6.1/dist/index.js";

const graphviz = await Graphviz.load();
const container = document.getElementById("graph-container");
const statusEl = document.getElementById("status");
const statusText = document.getElementById("status-text");
const sliderEl = document.getElementById("timeline-slider");
const liveBtnEl = document.getElementById("timeline-live");
const modeEl = document.getElementById("timeline-mode");
const metaEl = document.getElementById("timeline-meta");
const sourceEl = document.getElementById("snapshot-source");

let workingSnapshots = [];
let pastSnapshots = [];
let manualMode = false;
let selectedSnapshotID = null;

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
  container.innerHTML = '<p id="placeholder">No snapshots yet. Make file changes to build the session timeline.</p>';
}

function formatSnapshotMeta(snapshot, index, total) {
  const time = new Date(snapshot.timestamp);
  return `#${index + 1}/${total} | id ${snapshot.id} | ${time.toLocaleTimeString()}`;
}

function getAllSessionSnapshots() {
  return [...pastSnapshots, ...workingSnapshots];
}

function getHistoricalSnapshots(allSnapshots) {
  if (workingSnapshots.length === 0) {
    return allSnapshots;
  }
  const latestWorking = workingSnapshots[workingSnapshots.length - 1];
  return allSnapshots.filter((snapshot) => snapshot.id !== latestWorking.id);
}

function syncSnapshotSelector(allSnapshots, selectedSnapshot) {
  const selectedValue = selectedSnapshot ? `snapshot:${selectedSnapshot.id}` : "live";
  sourceEl.innerHTML = "";

  const liveOption = document.createElement("option");
  liveOption.value = "live";
  liveOption.textContent = "Live view";
  sourceEl.appendChild(liveOption);

  const historical = getHistoricalSnapshots(allSnapshots);
  historical.forEach((snapshot) => {
    const option = document.createElement("option");
    option.value = `snapshot:${snapshot.id}`;
    const time = new Date(snapshot.timestamp).toLocaleTimeString();
    option.textContent = `Snapshot ${snapshot.id} (${time})`;
    sourceEl.appendChild(option);
  });

  sourceEl.value = sourceEl.querySelector(`option[value="${selectedValue}"]`) ? selectedValue : "live";
}

function syncTimelineUI(allSnapshots, selectedIndex) {
  const total = allSnapshots.length;
  sliderEl.disabled = total <= 1;
  sliderEl.max = total > 0 ? String(total - 1) : "0";
  if (total === 0) {
    sliderEl.value = "0";
  } else {
    sliderEl.value = String(selectedIndex);
  }

  const snapshot = allSnapshots[selectedIndex];
  modeEl.textContent = manualMode ? "Session snapshot" : "Live view";
  liveBtnEl.disabled = !manualMode || total === 0;
  syncSnapshotSelector(allSnapshots, manualMode ? snapshot : null);

  if (snapshot) {
    metaEl.textContent = `${total} snapshots | ${formatSnapshotMeta(snapshot, selectedIndex, total)}`;
  } else {
    metaEl.textContent = "0 snapshots";
  }
}

function renderSelectedSnapshot() {
  const allSnapshots = getAllSessionSnapshots();
  if (allSnapshots.length === 0) {
    selectedSnapshotID = null;
    manualMode = false;
    syncTimelineUI(allSnapshots, 0);
    renderWaitingState();
    return;
  }

  let idx = allSnapshots.length - 1;
  if (manualMode && selectedSnapshotID !== null) {
    const selectedIdx = allSnapshots.findIndex((snapshot) => snapshot.id === selectedSnapshotID);
    if (selectedIdx >= 0) {
      idx = selectedIdx;
    } else {
      manualMode = false;
      selectedSnapshotID = null;
    }
  }

  const snapshot = allSnapshots[idx];
  if (!snapshot) {
    return;
  }
  renderGraph(snapshot.dot);
  syncTimelineUI(allSnapshots, idx);
}

function mergePayload(payload) {
  workingSnapshots = payload.workingSnapshots || [];
  pastSnapshots = payload.pastSnapshots || [];
  renderSelectedSnapshot();
}

sliderEl.addEventListener("input", function() {
  const allSnapshots = getAllSessionSnapshots();
  if (allSnapshots.length === 0) {
    return;
  }
  const idx = Math.min(Number(sliderEl.value || "0"), allSnapshots.length - 1);
  selectedSnapshotID = allSnapshots[idx].id;
  manualMode = true;
  renderSelectedSnapshot();
});

liveBtnEl.addEventListener("click", function() {
  manualMode = false;
  selectedSnapshotID = null;
  renderSelectedSnapshot();
});

sourceEl.addEventListener("change", function(event) {
  const selected = event.target.value;
  if (selected === "live") {
    manualMode = false;
    selectedSnapshotID = null;
    renderSelectedSnapshot();
    return;
  }

  if (!selected.startsWith("snapshot:")) {
    manualMode = false;
    selectedSnapshotID = null;
    renderSelectedSnapshot();
    return;
  }

  const id = Number(selected.split(":")[1]);
  if (!Number.isFinite(id)) {
    manualMode = false;
    selectedSnapshotID = null;
    renderSelectedSnapshot();
    return;
  }

  manualMode = true;
  selectedSnapshotID = id;
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
