console.log("app.js loaded");

let ws = null;
let token = localStorage.getItem("token");
let me = "";
let currentDialogId = null;
let dialogs = [];

const byId = (id) => document.getElementById(id);

function setStatus(text) {
    const el = byId("status");
    if (el) el.textContent = text;
}

/* ---------------- UI BIND ---------------- */

function bindUI() {
    const registerBtn = byId("registerBtn");
    const loginBtn = byId("loginBtn");
    const logoutBtn = byId("logoutBtn");
    const sendBtn = byId("sendBtn");
    const searchBtn = byId("searchBtn");

    registerBtn?.addEventListener("click", register);
    loginBtn?.addEventListener("click", login);
    logoutBtn?.addEventListener("click", logout);
    sendBtn?.addEventListener("click", sendMsg);
    searchBtn?.addEventListener("click", searchUsers);

    byId("msg")?.addEventListener("keydown", (e) => {
        if (e.key === "Enter") sendMsg();
    });

    byId("searchInput")?.addEventListener("keydown", (e) => {
        if (e.key === "Enter") searchUsers();
    });
}

/* ---------------- INIT ---------------- */

document.addEventListener("DOMContentLoaded", init);

async function init() {
    bindUI();

    if (!token) {
        showAuth();
        return;
    }

    const res = await fetch("/me", {
        headers: { "Authorization": token }
    });

    if (!res.ok) {
        localStorage.removeItem("token");
        token = null;
        showAuth();
        return;
    }

    const data = await res.json();
    me = data.username;

    byId("me").textContent = `@${me}`;

    showChat();
    await loadDialogs();
}

/* ---------------- UI ---------------- */

function showAuth() {
    byId("auth").classList.remove("hidden");
    byId("chatBox").classList.add("hidden");
}

function showChat() {
    byId("auth").classList.add("hidden");
    byId("chatBox").classList.remove("hidden");
}

/* ---------------- AUTH ---------------- */

async function register() {
    const res = await fetch("/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            username: byId("username").value,
            password: byId("password").value
        })
    });

    alert(await res.text());
}

async function login() {
    const res = await fetch("/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            username: byId("username").value,
            password: byId("password").value
        })
    });

    if (!res.ok) {
        alert(await res.text());
        return;
    }

    const data = await res.json();

    token = data.token;
    localStorage.setItem("token", token);

    const meRes = await fetch("/me", {
        headers: { "Authorization": token }
    });

    if (meRes.ok) {
        const meData = await meRes.json();
        me = meData.username;
        byId("me").textContent = `@${me}`;
    }

    showChat();
    await loadDialogs();
}

/* ---------------- DIALOGS ---------------- */

async function loadDialogs() {
    const res = await fetch("/dialogs", {
        headers: { "Authorization": token }
    });

    if (!res.ok) return;

    const data = await res.json();
    dialogs = Array.isArray(data) ? data : [];

    renderDialogs();

    if (!currentDialogId && dialogs.length > 0) {
        await openDialog(dialogs[0].id);
    }
}

function renderDialogs() {
    const list = byId("dialogs");
    if (!list) return;

    list.innerHTML = "";

    for (const dialog of dialogs) {
        const item = document.createElement("div");
        item.className = "dialog-item" + (dialog.id === currentDialogId ? " active" : "");
        item.onclick = () => openDialog(dialog.id);

        const title = document.createElement("div");
        title.textContent = dialog.title;

        const preview = document.createElement("div");
        preview.textContent = dialog.last_message || "No messages";

        item.appendChild(title);
        item.appendChild(preview);

        list.appendChild(item);
    }
}

/* ---------------- CHAT ---------------- */

async function openDialog(id) {
    currentDialogId = id;
    renderDialogs();

    closeWS();
    await loadMessages(id);
    connectWS(id);

    const dialog = dialogs.find(d => d.id === id);
    byId("chatHeader").textContent = dialog?.title || `Dialog #${id}`;
}

async function loadMessages(dialogId) {
    const res = await fetch(`/messages?dialog_id=${dialogId}`, {
        headers: { "Authorization": token }
    });

    if (!res.ok) return;

    const messages = await res.json();
    const chat = byId("chat");

    chat.innerHTML = "";

    for (const msg of messages) {
        addMessage(msg);
    }

    chat.scrollTop = chat.scrollHeight;
}

function addMessage(msg) {
    const chat = byId("chat");

    const div = document.createElement("div");
    div.className = "message" + (msg.username === me ? " me" : "");

    div.innerHTML = `
        <span class="username">${msg.username}</span>
        <span class="text">${msg.text}</span>
    `;

    chat.appendChild(div);
    chat.scrollTop = chat.scrollHeight;
}

/* ---------------- WS ---------------- */

function connectWS(dialogId) {
    const protocol = location.protocol === "https:" ? "wss" : "ws";

    ws = new WebSocket(
        `${protocol}://${location.host}/ws?token=${encodeURIComponent(token)}&dialog_id=${dialogId}`
    );

    ws.onmessage = (e) => {
        const msg = JSON.parse(e.data);

        if (msg.dialog_id !== currentDialogId) return;

        addMessage(msg);
    };

    ws.onclose = () => {
        ws = null;
    };
}

function closeWS() {
    if (ws) {
        ws.close();
        ws = null;
    }
}

/* ---------------- SEND ---------------- */

function sendMsg() {
    if (!ws || !currentDialogId) return;

    const input = byId("msg");
    const text = input.value.trim();

    if (!text) return;

    ws.send(JSON.stringify({ text }));
    input.value = "";
}

/* ---------------- SEARCH ---------------- */

async function searchUsers() {
    const q = byId("searchInput").value.trim();
    const results = byId("searchResults");

    if (!q) {
        results.innerHTML = "";
        return;
    }

    const res = await fetch(`/users/search?q=${encodeURIComponent(q)}`, {
        headers: { "Authorization": token }
    });

    if (!res.ok) return;

    const users = await res.json();

    results.innerHTML = "";

    for (const u of users) {
        const div = document.createElement("div");
        div.className = "user-item";
        div.textContent = u;
        div.onclick = () => createDialog(u);
        results.appendChild(div);
    }
}

async function createDialog(username) {
    const res = await fetch("/dialogs/create", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": token
        },
        body: JSON.stringify({ username })
    });

    if (!res.ok) {
        alert(await res.text());
        return;
    }

    const data = await res.json();

    await loadDialogs();
    await openDialog(data.dialog_id);
}

/* ---------------- LOGOUT ---------------- */

function logout() {
    localStorage.removeItem("token");

    token = null;
    me = "";
    currentDialogId = null;

    closeWS();
    showAuth();
}