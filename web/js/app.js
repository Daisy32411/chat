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

function bindUI() {
    byId("registerBtn").addEventListener("click", register);
    byId("loginBtn").addEventListener("click", login);
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

async function init() {
    bindUI();

    if (!token) {
        showAuth();
        setStatus("");
        return;
    }

    try {
        const res = await fetch("/me", {
            headers: { "Authorization": token }
        });

        if (!res.ok) {
            localStorage.removeItem("token");
            token = null;
            showAuth();
            setStatus("");
            return;
        }

        const data = await res.json();
        me = data.username;
        byId("me").textContent = `@${me}`;

        showChat();
        await loadDialogs();
    } catch (err) {
        console.error(err);
        setStatus("Network error");
        showAuth();
    }
}

function showAuth() {
    byId("auth").classList.remove("hidden");
    byId("chatBox").classList.add("hidden");
}

function showChat() {
    byId("auth").classList.add("hidden");
    byId("chatBox").classList.remove("hidden");
}

async function register() {
    console.log("register click");
    setStatus("Registering...");

    try {
        const res = await fetch("/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                username: byId("username").value,
                password: byId("password").value
            })
        });

        const text = await res.text();

        if (!res.ok) {
            setStatus(text || "Register failed");
            alert(text || "Register failed");
            return;
        }

        setStatus("Registered");
        alert(text || "registered");
    } catch (err) {
        console.error(err);
        setStatus("Register network error");
        alert("Register network error");
    }
}

async function login() {
    console.log("login click");
    setStatus("Logging in...");

    try {
        const res = await fetch("/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                username: byId("username").value,
                password: byId("password").value
            })
        });

        if (!res.ok) {
            const text = await res.text();
            setStatus(text || "Invalid login");
            alert(text || "Invalid login");
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
        setStatus("");
    } catch (err) {
        console.error(err);
        setStatus("Login network error");
        alert("Login network error");
    }
}

async function loadDialogs() {
    try {
        const res = await fetch("/dialogs", {
            headers: { "Authorization": token }
        });

        if (!res.ok) {
            dialogs = [];
            renderDialogs();
            return;
        }

        const data = await res.json();
        dialogs = Array.isArray(data) ? data : [];
        renderDialogs();

        if (!currentDialogId && dialogs.length > 0) {
            await openDialog(dialogs[0].id);
        }

        if (dialogs.length === 0) {
            byId("chatHeader").textContent = "Select a dialog";
            byId("chat").innerHTML = "";
        }
    } catch (err) {
        console.error(err);
        dialogs = [];
        renderDialogs();
    }
}

function renderDialogs() {
    const list = byId("dialogs");
    list.innerHTML = "";

    for (const dialog of dialogs) {
        const item = document.createElement("div");
        item.className = "dialog-item" + (dialog.id === currentDialogId ? " active" : "");
        item.onclick = () => openDialog(dialog.id);

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
    if (!ws) return;
    const socket = ws;
    ws = null;
    socket.close();
}

async function openDialog(id) {
    currentDialogId = id;
    renderDialogs();

    closeWS();
    await loadMessages(id);
    connectWS(id);

    const dialog = dialogs.find(d => d.id === id);
    byId("chatHeader").textContent = dialog ? dialog.title : `Dialog #${id}`;
}

async function loadMessages(dialogId) {
    try {
        const res = await fetch(`/messages?dialog_id=${dialogId}`, {
            headers: { "Authorization": token }
        });

        if (!res.ok) {
            return;
        }

        const messages = await res.json();
        const chat = byId("chat");
        chat.innerHTML = "";

        for (const msg of messages) {
            addMessage(msg);
        }

        chat.scrollTop = chat.scrollHeight;
    } catch (err) {
        console.error(err);
    }
}

function connectWS(dialogId) {
    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    const socket = new WebSocket(
        `${protocol}://${window.location.host}/ws?token=${encodeURIComponent(token)}&dialog_id=${dialogId}`
    );

    ws = socket;

    socket.onmessage = (e) => {
        const msg = JSON.parse(e.data);
        if (msg.dialog_id !== currentDialogId) return;
        addMessage(msg);
    };

    socket.onclose = () => {
        if (ws === socket) ws = null;
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

    try {
        const res = await fetch(`/users/search?q=${encodeURIComponent(q)}`, {
            headers: { "Authorization": token }
        });

        if (!res.ok) {
            results.innerHTML = "";
            return;
        }

        const users = await res.json();
        results.innerHTML = "";

        for (const username of users) {
            const item = document.createElement("div");
            item.className = "user-item";
            item.textContent = username;
            item.onclick = () => createDialog(username);
            results.appendChild(item);
        }
    } catch (err) {
        console.error(err);
    }
}

async function createDialog(username) {
    try {
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
        await openDialog(data.dialog_id);
    } catch (err) {
        console.error(err);
    }
}

function logout() {
    localStorage.removeItem("token");
    token = null;
    me = "";
    currentDialogId = null;

    closeWS();
    showAuth();

    byId("me").textContent = "";
    byId("dialogs").innerHTML = "";
    byId("searchResults").innerHTML = "";
    byId("chat").innerHTML = "";
    byId("chatHeader").textContent = "Select a dialog";
    setStatus("");
}

init();