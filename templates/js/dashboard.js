let ws = null;
let currentChat = null;
const currentUser = document.body.getAttribute('data-username');

function connectWebSocket() {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    ws = new WebSocket(protocol + "//" + window.location.host + "/ws");
    ws.onopen = () => console.log("WebSocket connected");
    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (currentChat && (data.from === currentChat || data.to === currentChat)) {
            appendMessage(data.from, data.text, data.from === currentUser ? "sent" : "received");
        } else {
            loadDialogs();
        }
    };
    ws.onerror = (error) => console.error("WebSocket error", error);
    ws.onclose = () => {
        console.log("WebSocket closed, reconnecting...");
        setTimeout(connectWebSocket, 3000);
    };
}

function loadDialogs() {
    fetch("/api/dialogs")
        .then(res => res.json())
        .then(dialogs => {
            const container = document.getElementById("dialogsList");
            container.innerHTML = "";
            dialogs.forEach(d => {
                const div = document.createElement("div");
                div.className = "dialog-item";
                if (currentChat === d.with_user) div.classList.add("active");
                div.innerHTML = `<div class="dialog-name">${escapeHtml(d.with_user)}</div>
                                 <div class="dialog-lastmsg">${escapeHtml(d.last_message.substring(0, 50))}</div>`;
                div.onclick = () => openChat(d.with_user);
                container.appendChild(div);
            });
        });
}

function openChat(username) {
    currentChat = username;
    document.getElementById("chatHeader").innerText = `Чат с ${username}`;
    document.getElementById("inputArea").style.display = "flex";
    fetch(`/api/messages?with=${encodeURIComponent(username)}`)
        .then(res => res.json())
        .then(messages => {
            const container = document.getElementById("messagesContainer");
            container.innerHTML = "";
            messages.forEach(msg => {
                const dir = msg.from_user === currentUser ? "sent" : "received";
                appendMessage(msg.from_user, msg.text, dir);
            });
            container.scrollTop = container.scrollHeight;
        });
    loadDialogs();
}

function appendMessage(sender, text, direction) {
    const container = document.getElementById("messagesContainer");
    const div = document.createElement("div");
    div.className = `message ${direction}`;
    div.innerHTML = `<div class="message-bubble">${escapeHtml(text)}</div>
                     <div class="message-info">${escapeHtml(sender)}</div>`;
    container.appendChild(div);
    container.scrollTop = container.scrollHeight;
}

function sendMessage() {
    if (!currentChat) return;
    const input = document.getElementById("messageInput");
    const text = input.value.trim();
    if (!text) return;
    ws.send(JSON.stringify({ to: currentChat, text: text }));
    input.value = "";
}

function escapeHtml(str) {
    return str.replace(/[&<>]/g, function(m) {
        if (m === '&') return '&amp;';
        if (m === '<') return '&lt;';
        if (m === '>') return '&gt;';
        return m;
    });
}

let searchTimeout;
document.getElementById("searchInput").addEventListener("input", function(e) {
    clearTimeout(searchTimeout);
    const query = e.target.value.trim();
    const resultsDiv = document.getElementById("searchResults");
    if (query.length < 2) {
        resultsDiv.style.display = "none";
        return;
    }
    searchTimeout = setTimeout(() => {
        fetch(`/api/search?q=${encodeURIComponent(query)}`)
            .then(res => res.json())
            .then(users => {
                if (users.length === 0) {
                    resultsDiv.style.display = "none";
                    return;
                }
                resultsDiv.innerHTML = "";
                users.forEach(u => {
                    const item = document.createElement("div");
                    item.className = "search-result-item";
                    item.textContent = u;
                    item.onclick = () => {
                        openChat(u);
                        resultsDiv.style.display = "none";
                        document.getElementById("searchInput").value = "";
                    };
                    resultsDiv.appendChild(item);
                });
                resultsDiv.style.display = "block";
            });
    }, 300);
});
document.addEventListener("click", function(e) {
    if (!e.target.closest(".search-box")) {
        document.getElementById("searchResults").style.display = "none";
    }
});

document.getElementById("sendBtn").addEventListener("click", sendMessage);
document.getElementById("messageInput").addEventListener("keypress", function(e) {
    if (e.key === "Enter") sendMessage();
});

connectWebSocket();
loadDialogs();