
let events = [];
let feedInterval = 210 * 60 * 1000;
let sleepInterval = 120 * 60 * 1000;
let telegramId = null;
let babyId = null;

window.onload = async () => {
  const tg = window.Telegram?.WebApp;
  telegramId = tg?.initDataUnsafe?.user?.id || 123456;

  const authRes = await fetch("/auth", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ telegram_id: telegramId })
  });
  const authData = await authRes.json();
  babyId = authData.baby_id;
  await loadEvents();
  updateNextTimes();
};

async function loadEvents() {
  const res = await fetch(`/events?telegram_id=${telegramId}`);
  const data = await res.json();
  events = data;
  renderEvents();
}

async function logEvent(type) {
  const now = new Date();
  const timeStr = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  const timestamp = now.getTime();

  await fetch("/events", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ telegram_id: telegramId, type, time_str: timeStr, timestamp })
  });

  await loadEvents();
  updateNextTimes();
}

function renderEvents() {
  const list = document.getElementById("eventHistory");
  list.innerHTML = '';
  const sorted = [...events].sort((a, b) => b.timestamp - a.timestamp);
  sorted.forEach(event => {
    const item = document.createElement("li");
    item.className = "event-item";
    const span = document.createElement("span");
    span.textContent = `${event.type} — ${event.time_str}`;
    const remove = document.createElement("button");
    remove.textContent = "✖";
    remove.className = "remove-button";
    remove.onclick = () => deleteEvent(event.id);
    item.appendChild(span);
    item.appendChild(remove);
    list.appendChild(item);
  });
}

function deleteEvent(id) {
  events = events.filter(e => e.id !== id);
  renderEvents();
}

function formatMs(ms) {
  const m = Math.floor(ms / 60000);
  const h = Math.floor(m / 60);
  const rem = m % 60;
  return h > 0 ? `${h} ч ${rem} мин` : `${rem} мин`;
}

function updateNextTimes() {
  const now = Date.now();
  const lastFeed = [...events].filter(e => e.type === "Кормление").sort((a, b) => b.timestamp - a.timestamp)[0];
  const lastWake = [...events].filter(e => e.type === "Проснулся").sort((a, b) => b.timestamp - a.timestamp)[0];

  const nextFeed = lastFeed ? lastFeed.timestamp + feedInterval : null;
  const nextSleep = lastWake ? lastWake.timestamp + sleepInterval : null;

  document.getElementById("nextFeed").textContent = nextFeed ?
    (nextFeed > now ? formatMs(nextFeed - now) : "уже пора!") : "–";

  document.getElementById("nextSleep").textContent = nextSleep ?
    (nextSleep > now ? formatMs(nextSleep - now) : "уже пора!") : "–";
}

function openModal(id) {
  document.getElementById(id).classList.remove("hidden");
}
function closeModal(id) {
  document.getElementById(id).classList.add("hidden");
}

function saveManualEvent() {
  const type = document.getElementById("manualType").value;
  const time = document.getElementById("manualTime").value;
  if (!time) return;
  const [h, m] = time.split(":");
  const now = new Date();
  now.setHours(h, m, 0, 0);
  const timestamp = now.getTime();
  fetch("/events", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ telegram_id: telegramId, type, time_str: time, timestamp })
  }).then(() => {
    closeModal("manualModal");
    loadEvents();
    updateNextTimes();
  });
}

function saveIntervalSettings() {
  const feedMin = parseInt(document.getElementById("feedIntervalInput").value);
  const sleepMin = parseInt(document.getElementById("sleepIntervalInput").value);
  if (!isNaN(feedMin)) feedInterval = feedMin * 60000;
  if (!isNaN(sleepMin)) sleepInterval = sleepMin * 60000;
  closeModal("settingsModal");
  updateNextTimes();
}

async function generateInvite() {
  const res = await fetch("/invite", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ telegram_id: telegramId })
  });
  const data = await res.json();
  document.getElementById("inviteCode").textContent = data.code || "–";
}

async function joinFamily() {
  const code = document.getElementById("inviteInput").value.trim();
  if (!code) return;
  await fetch("/join", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ telegram_id: telegramId, code })
  });
  closeModal("inviteModal");
  await loadEvents();
  updateNextTimes();
}
