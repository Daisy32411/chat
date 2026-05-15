let ws = null;
let token = localStorage.getItem("token");

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

    showChat();
}

init();

function showChat() {
    document.getElementById("auth").classList.add("hidden");
    document.getElementById("chatBox").classList.remove("hidden");

    loadMessages();
    connectWS();
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
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            username: document.getElementById("username").value,
            password: document.getElementById("password").value
        })
    });

    if (!res.ok) {
        alert("Invalid login");
        return;
    }

    const data = await res.json();

    token = data.token;

    localStorage.setItem("token", token);

    showChat();
}

function connectWS() {
    if (ws) return;

    ws = new WebSocket(
        "ws://77.223.101.245:8080/ws?token=" + token
    );

    ws.onmessage = (e) => {
        const msg = JSON.parse(e.data);

        const div = document.createElement("div");

        div.innerHTML = `<b>${msg.username}</b>: ${msg.text}`;

        document.getElementById("chat").appendChild(div);

        document.getElementById("chat").scrollTop =
            document.getElementById("chat").scrollHeight;
    };

    ws.onclose = () => {
        console.log("WS closed");
        ws = null;
    };
}

function sendMsg() {
    if (!ws) return;

    const input = document.getElementById("msg");

    ws.send(JSON.stringify({
        dialog_id: 1,
        text: input.value
    }));

    input.value = "";
}

function logout() {
    localStorage.removeItem("token");

    token = null;

    if (ws) {
        ws.close();
    }

    ws = null;

    showAuth();
}

async function loadMessages() {
    const res = await fetch("/messages?dialog_id=1");

    const messages = await res.json();

    const chat = document.getElementById("chat");

    chat.innerHTML = "";

    for (const msg of messages) {
        const div = document.createElement("div");

        div.innerHTML = `<b>${msg.username}</b>: ${msg.text}`;

        chat.appendChild(div);
    }

    chat.scrollTop = chat.scrollHeight;
}