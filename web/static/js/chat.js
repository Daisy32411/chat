const meEl = document.getElementById('me');
const dialogsEl = document.getElementById('dialogs');
const searchInput = document.getElementById('user-search');
const searchResultsEl = document.getElementById('search-results');
const messagesEl = document.getElementById('messages');
const chatTitleEl = document.getElementById('chat-title');
const chatSubtitleEl = document.getElementById('chat-subtitle');
const sendForm = document.getElementById('send-form');
const messageInput = document.getElementById('message-input');
const logoutBtn = document.getElementById('logout-btn');

let me = null;
let currentDialogId = null;
let currentDialogUser = null;
let dialogs = [];
let refreshTimer = null;

async function api(url, options = {}) {
  const res = await fetch(url, {
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || 'Request failed');
  return data;
}

function formatTime(value) {
  return new Date(value).toLocaleString([], {
    hour: '2-digit',
    minute: '2-digit',
    day: '2-digit',
    month: '2-digit',
  });
}

function renderDialogs() {
  dialogsEl.innerHTML = '';
  dialogs.forEach((d) => {
    const last = d.last_message ? d.last_message.text : 'No messages yet';
    const node = document.createElement('div');
    node.className = 'item' + (d.dialog_id === currentDialogId ? ' active' : '');
    node.innerHTML = `
      <div class="item-title">${d.other_user.username}</div>
      <div class="item-subtitle">${last}</div>
    `;
    node.addEventListener('click', async () => {
      await openDialogById(d.dialog_id, d.other_user);
    });
    dialogsEl.appendChild(node);
  });
}

function renderMessages(messages) {
  messagesEl.innerHTML = '';
  messages.forEach(renderMessage);
  messagesEl.scrollTop = messagesEl.scrollHeight;
}

function renderMessage(message) {
  const node = document.createElement('div');
  node.className = 'msg' + (message.sender_id === me.id ? ' me' : '');
  node.innerHTML = `
    <div class="msg-meta">${formatTime(message.created_at)}</div>
    <div>${escapeHtml(message.text)}</div>
  `;
  messagesEl.appendChild(node);
  messagesEl.scrollTop = messagesEl.scrollHeight;
}

function escapeHtml(str) {
  return String(str)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

async function loadMe() {
  const data = await api('/api/me');
  me = data.user;
  meEl.textContent = `@${me.username}`;
}

async function loadDialogs() {
  const data = await api('/api/dialogs');
  dialogs = data.dialogs || [];
  renderDialogs();
}

async function loadMessages(dialogId) {
  const data = await api(`/api/messages?dialog_id=${dialogId}`);
  renderMessages(data.messages || []);
}

async function openDialogById(dialogId, otherUser = null) {
  currentDialogId = dialogId;
  currentDialogUser = otherUser || null;

  if (!currentDialogUser) {
    const d = dialogs.find(x => x.dialog_id === dialogId);
    if (d) currentDialogUser = d.other_user;
  }

  chatTitleEl.textContent = currentDialogUser ? currentDialogUser.username : `Dialog #${dialogId}`;
  chatSubtitleEl.textContent = currentDialogUser ? currentDialogUser.email : 'Chat';
  await loadMessages(dialogId);
  renderDialogs();
  startPolling();
}

function startPolling() {
  if (refreshTimer) clearInterval(refreshTimer);
  refreshTimer = setInterval(async () => {
    if (!currentDialogId) return;
    try {
      await loadMessages(currentDialogId);
      await loadDialogs();
    } catch {}
  }, 2500);
}

async function searchUsers(query) {
  const data = await api(`/api/users/search?q=${encodeURIComponent(query)}`);
  searchResultsEl.innerHTML = '';
  (data.users || []).forEach((user) => {
    if (user.id === me.id) return;
    const node = document.createElement('div');
    node.className = 'item';
    node.innerHTML = `
      <div class="item-title">${user.username}</div>
      <div class="item-subtitle">${user.email}</div>
    `;
    node.addEventListener('click', async () => {
      const open = await api('/api/dialogs/open', {
        method: 'POST',
        body: JSON.stringify({ user_id: user.id }),
      });
      await loadDialogs();
      const dialogId = open.dialog_id;
      await openDialogById(dialogId, user);
      searchInput.value = '';
      searchResultsEl.innerHTML = '';
    });
    searchResultsEl.appendChild(node);
  });
}

searchInput.addEventListener('input', async () => {
  const q = searchInput.value.trim();
  if (!q) {
    searchResultsEl.innerHTML = '';
    return;
  }
  try {
    await searchUsers(q);
  } catch (err) {
    searchResultsEl.innerHTML = `<div class="item"><div class="item-subtitle">${escapeHtml(err.message)}</div></div>`;
  }
});

sendForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  if (!currentDialogId) return;
  const text = messageInput.value.trim();
  if (!text) return;

  messageInput.value = '';
  try {
    await api('/api/messages/send', {
      method: 'POST',
      body: JSON.stringify({ dialog_id: currentDialogId, text }),
    });
    await loadMessages(currentDialogId);
    await loadDialogs();
  } catch (err) {
    alert(err.message);
  }
});

logoutBtn.addEventListener('click', async () => {
  await api('/api/auth/logout', { method: 'POST', body: '{}' });
  window.location.href = '/login';
});

(async function init() {
  try {
    await loadMe();
    await loadDialogs();
    if (dialogs.length > 0) {
      await openDialogById(dialogs[0].dialog_id, dialogs[0].other_user);
    } else {
      chatSubtitleEl.textContent = 'Search a user to start chatting.';
      startPolling();
    }
  } catch (err) {
    window.location.href = '/login';
  }
})();
