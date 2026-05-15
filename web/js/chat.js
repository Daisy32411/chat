console.log("chat.js loaded");

let ws = null;
let token = localStorage.getItem("token");
let me = "";
let currentDialogId = null;
let dialogs = [];

const byId = (id) => document.getElementById(id);

function ensureToken() {
    if (!token) {
        window.location.href = "/";
        return false;
    }
    return true;
}

function bindUI() {
    byId("logoutBtn").addEventListener("click", logout);
    byId("sendBtn").addEventListener("click", sendMsg);
    byId("searchBtn").addEventListener("click", searchUsers);

    byId("msg").addEventListener("keydown", (e) => {
        if (e.key === "Enter") sendMsg();
    });

    byId("searchInput").addEventListener("keydown", (e) => {
        if (e.key === "Enter") searchUsers();
    });
}

document.addEventListener("DOMContentLoaded", init);

async function init() {
    bindUI();

    if (!ensureToken()) return;

    const res = await fetch("/me", {
        headers: { "Authorization": token }
    });

    if (!res.ok) {
        localStorage.removeItem("token");
        token = null;
        window.location.href = "/";
        return;
    }

    const data = await res.json();
    me = data.username;
    byId("me").textContent = `@${me}`;

    await loadDialogs();
}

async function loadDialogs() {
    const res = await fetch("/dialogs", {
        headers: { "Authorization": token }
    });

    if (!res.ok) {
        if (res.status === 401) {
            logout();
            return;
        }

        dialogs = [];
        renderDialogs();
        return;
    }

    const data = await res.json();
    dialogs = Array.isArray(data) ? data : [];
    renderDialogs();

    if (dialogs.length > 0 && !currentDialogId) {
        await openDialog(dialogs[0]);
    } else if (dialogs.length === 0) {
        byId("chatHeader").textContent = "Select a dialog";
        byId("chat").innerHTML = "";
    }
}

function renderDialogs() {
    const list = byId("dialogs");
    list.innerHTML = "";

    if (dialogs.length === 0) {
        const empty = document.createElement("div");
        empty.className = "dialog-preview";
        empty.textContent = "No chats yet. Search users to start.";
        list.appendChild(empty);
        return;
    }

    for (const dialog of dialogs) {
        const item = document.createElement("div");
        item.className = "dialog-item" + (dialog.id === currentDialogId ? " active" : "");
        item.onclick = () => openDialog(dialog);

        const title = document.createElement("div");
        title.className = "dialog-title";
        title.textContent = dialog.title;

        const preview = document.createElement("div");
        preview.className = "dialog-preview";
        preview.textContent = dialog.last_message || "No messages yet";

        item.appendChild(title);
        item.appendChild(preview);
        list.appendChild(item);
    }
}

function closeWS() {
    if (ws) {
        ws.close();
        ws = null;
    }
}

async function openDialog(dialog) {
    currentDialogId = dialog.id;
    renderDialogs();
    byId("chatHeader").textContent = dialog.title;

    closeWS();

    await loadMessages(dialog.id);
    connectWS(dialog.id);
}

async function loadMessages(dialogId) {
    const res = await fetch(`/messages?dialog_id=${dialogId}`, {
        headers: { "Authorization": token }
    });

    if (!res.ok) {
        if (res.status === 401) {
            logout();
        }
        return;
    }

    const messages = await res.json();
    const chat = byId("chat");
    chat.innerHTML = "";

    for (const msg of messages) {
        addMessage(msg);
    }

    chat.scrollTop = chat.scrollHeight;
}

function connectWS(dialogId) {
    const protocol = location.protocol === "https:" ? "wss" : "ws";
    ws = new WebSocket(
        `${protocol}://${location.host}/ws?token=${encodeURIComponent(token)}&dialog_id=${dialogId}`
    );

    ws.onmessage = (e) => {
        const msg = JSON.parse(e.data);
        addMessage(msg);
        updateDialogPreview(msg);
    };

    ws.onclose = () => {
        ws = null;
    };

    ws.onerror = (e) => {
        console.error("WS error", e);
    };
}

function addMessage(msg) {
    const chat = byId("chat");
    const div = document.createElement("div");
    div.className = "message" + (msg.username === me ? " me" : "");

    const name = document.createElement("span");
    name.className = "username";
    name.textContent = msg.username;

    const text = document.createElement("span");
    text.className = "text";
    text.textContent = msg.text;

    div.appendChild(name);
    div.appendChild(text);

    chat.appendChild(div);
    chat.scrollTop = chat.scrollHeight;
}

function updateDialogPreview(msg) {
    const dialog = dialogs.find(d => d.id === currentDialogId);
    if (!dialog) return;

    dialog.last_message = msg.text;
    renderDialogs();
}

function sendMsg() {
    if (!ws || !currentDialogId) return;

    const input = byId("msg");
    const text = input.value.trim();
    if (!text) return;

    ws.send(JSON.stringify({ text }));
    input.value = "";
}

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

    if (!users.length) {
        const empty = document.createElement("div");
        empty.className = "dialog-preview";
        empty.textContent = "No users found";
        results.appendChild(empty);
        return;
    }

    for (const username of users) {
        if (username === me) continue;

        const item = document.createElement("div");
        item.className = "user-item";
        item.textContent = username;
        item.onclick = () => createDialog(username);
        results.appendChild(item);
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

    byId("searchInput").value = "";
    byId("searchResults").innerHTML = "";

    await loadDialogs();

    const dialog = dialogs.find(d => d.id === data.dialog_id);
    if (dialog) {
        await openDialog(dialog);
    }
}

function logout() {
    localStorage.removeItem("token");
    token = null;
    me = "";
    currentDialogId = null;
    dialogs = [];

    closeWS();
    window.location.href = "/";
}