console.log("chat.js loaded");

let ws = null;
let token = localStorage.getItem("token");
let me = "";
let currentDialogId = null;
let dialogs = [];

const byId = (id) => document.getElementById(id);

document.addEventListener("DOMContentLoaded", init);

async function init() {
    if (!token) return redirectAuth();

    bindUI();

    const meRes = await fetch("/me", {
        headers: { Authorization: token }
    });

    if (!meRes.ok) return redirectAuth();

    const meData = await meRes.json();
    me = meData.username;
    byId("me").textContent = `@${me}`;

    await loadDialogs();
}

function redirectAuth() {
    localStorage.removeItem("token");
    window.location.href = "/";
}

function bindUI() {
    byId("logoutBtn").onclick = logout;
    byId("sendBtn").onclick = sendMsg;
    byId("searchBtn").onclick = searchUsers;

    byId("msg").addEventListener("keydown", e => {
        if (e.key === "Enter") sendMsg();
    });

    byId("searchInput").addEventListener("keydown", e => {
        if (e.key === "Enter") searchUsers();
    });
}

/* ================= DIALOGS ================= */

async function loadDialogs() {
    const res = await fetch("/dialogs", {
        headers: { Authorization: token }
    });

    if (!res.ok) {
        dialogs = [];
        renderDialogs();
        return;
    }

    dialogs = await res.json();

    renderDialogs();

    if (dialogs.length > 0 && !currentDialogId) {
        openDialog(dialogs[0].id);
    }
}

function renderDialogs() {
    const list = byId("dialogs");
    list.innerHTML = "";

    for (const d of dialogs) {
        const el = document.createElement("div");
        el.className = "dialog-item" + (d.id === currentDialogId ? " active" : "");

        el.onclick = () => openDialog(d.id);

        el.innerHTML = `
            <div class="dialog-title">${d.title}</div>
            <div class="dialog-preview">${d.last_message || ""}</div>
        `;

        list.appendChild(el);
    }
}

/* ================= CHAT ================= */

async function openDialog(id) {
    currentDialogId = id;

    renderDialogs();
    closeWS();

    await loadMessages(id);
    connectWS(id);

    const d = dialogs.find(x => x.id === id);
    byId("chatHeader").textContent = d ? d.title : "Chat";
}

async function loadMessages(dialogId) {
    const res = await fetch(`/messages?dialog_id=${dialogId}`, {
        headers: { Authorization: token }
    });

    if (!res.ok) return;

    const messages = await res.json();

    const chat = byId("chat");
    chat.innerHTML = "";

    for (const m of messages) {
        drawMessage(m);
    }

    chat.scrollTop = chat.scrollHeight;
}

function connectWS(dialogId) {
    const proto = location.protocol === "https:" ? "wss" : "ws";

    ws = new WebSocket(
        `${proto}://${location.host}/ws?token=${encodeURIComponent(token)}&dialog_id=${dialogId}`
    );

    ws.onmessage = (e) => {
        const msg = JSON.parse(e.data);

        if (msg.dialog_id !== currentDialogId) return;

        drawMessage(msg);
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

function drawMessage(msg) {
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

/* ================= SEND ================= */

function sendMsg() {
    if (!ws || !currentDialogId) return;

    const input = byId("msg");
    const text = input.value.trim();
    if (!text) return;

    ws.send(JSON.stringify({ text }));
    input.value = "";
}

/* ================= SEARCH ================= */

async function searchUsers() {
    const q = byId("searchInput").value.trim();
    const resBox = byId("searchResults");

    if (!q) {
        resBox.innerHTML = "";
        return;
    }

    const res = await fetch(`/users/search?q=${encodeURIComponent(q)}`, {
        headers: { Authorization: token }
    });

    if (!res.ok) return;

    const users = await res.json();

    resBox.innerHTML = "";

    for (const u of users) {
        const el = document.createElement("div");
        el.className = "user-item";
        el.textContent = u;

        el.onclick = () => createDialog(u);

        resBox.appendChild(el);
    }
}

async function createDialog(username) {
    const res = await fetch("/dialogs/create", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Authorization: token
        },
        body: JSON.stringify({ username })
    });

    if (!res.ok) return alert(await res.text());

    const data = await res.json();

    await loadDialogs();
    openDialog(data.dialog_id);
}

/* ================= LOGOUT ================= */

function logout() {
    localStorage.removeItem("token");
    window.location.href = "/";
}