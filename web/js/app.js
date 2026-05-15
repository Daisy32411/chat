let ws = null;
let token = localStorage.getItem("token");
let me = "";
let currentDialogId = null;
let dialogs = [];

async function init() {
    if (!token) {
        showAuth();
        return;
    }

    const res = await fetch("/me", {
        headers: {
            "Authorization": token
        }
    });

    if (!res.ok) {
        localStorage.removeItem("token");
        token = null;
        showAuth();
        return;
    }

    const data = await res.json();
    me = data.username;
    document.getElementById("me").textContent = `@${me}`;

    showChat();
    await loadDialogs();
}

function showChat() {
    document.getElementById("auth").classList.add("hidden");
    document.getElementById("chatBox").classList.remove("hidden");
}

function showAuth() {
    document.getElementById("auth").classList.remove("hidden");
    document.getElementById("chatBox").classList.add("hidden");
}

async function register() {
    const res = await fetch("/register", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            username: document.getElementById("username").value,
            password: document.getElementById("password").value
        })
    });

    alert(await res.text());
}

async function login() {
    const res = await fetch("/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            username: document.getElementById("username").value,
            password: document.getElementById("password").value
        })
    });

    if (!res.ok) {
        alert(await res.text());
        return;
    }

    const data = await res.json();

    token = data.token;
    localStorage.setItem("token", token);

    showChat();
}

async function loadDialogs() {
    const res = await fetch("/dialogs", {
        headers: {
            "Authorization": token
        }
    });

    if (!res.ok) {
        return;
    }

    dialogs = await res.json();
    renderDialogs();

    if (!currentDialogId && dialogs.length > 0) {
        await openDialog(dialogs[0].id);
    }

    if (dialogs.length === 0) {
        document.getElementById("chatHeader").textContent = "Select a dialog";
        document.getElementById("chat").innerHTML = "";
    }
}

function renderDialogs() {
    const list = document.getElementById("dialogs");
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
    if (!ws) {
        return;
    }

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
    document.getElementById("chatHeader").textContent = dialog ? dialog.title : `Dialog #${id}`;
}

async function loadMessages(dialogId) {
    const res = await fetch(`/messages?dialog_id=${dialogId}`, {
        headers: {
            "Authorization": token
        }
    });

    if (!res.ok) {
        return;
    }

    const messages = await res.json();
    const chat = document.getElementById("chat");
    chat.innerHTML = "";

    for (const msg of messages) {
        addMessage(msg);
    }

    chat.scrollTop = chat.scrollHeight;
}

function connectWS(dialogId) {
    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    const socket = new WebSocket(
        `${protocol}://${window.location.host}/ws?token=${encodeURIComponent(token)}&dialog_id=${dialogId}`
    );

    ws = socket;

    socket.onmessage = (e) => {
        const msg = JSON.parse(e.data);

        if (msg.dialog_id !== currentDialogId) {
            return;
        }

        addMessage(msg);
    };

    socket.onclose = () => {
        if (ws === socket) {
            ws = null;
        }
    };
}

function addMessage(msg) {
    const chat = document.getElementById("chat");
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
    if (!ws || !currentDialogId) {
        return;
    }

    const input = document.getElementById("msg");
    const text = input.value.trim();

    if (!text) {
        return;
    }

    ws.send(JSON.stringify({
        text: text
    }));

    input.value = "";
}

async function searchUsers() {
    const q = document.getElementById("searchInput").value.trim();
    const results = document.getElementById("searchResults");

    if (!q) {
        results.innerHTML = "";
        return;
    }

    const res = await fetch(`/users/search?q=${encodeURIComponent(q)}`, {
        headers: {
            "Authorization": token
        }
    });

    if (!res.ok) {
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
}

async function createDialog(username) {
    const res = await fetch("/dialogs/create", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": token
        },
        body: JSON.stringify({
            username: username
        })
    });

    if (!res.ok) {
        alert(await res.text());
        return;
    }

    const data = await res.json();
    document.getElementById("searchInput").value = "";
    document.getElementById("searchResults").innerHTML = "";

    await loadDialogs();
    await openDialog(data.dialog_id);
}

function logout() {
    localStorage.removeItem("token");
    token = null;
    me = "";
    currentDialogId = null;

    closeWS();
    showAuth();
    document.getElementById("me").textContent = "";
    document.getElementById("dialogs").innerHTML = "";
    document.getElementById("searchResults").innerHTML = "";
    document.getElementById("chat").innerHTML = "";
    document.getElementById("chatHeader").textContent = "Select a dialog";
}

init();