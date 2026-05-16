let ws = null;
let currentChat = null;
const currentUser = document.body.getAttribute('data-username');

// Профильная панель
const menuToggle = document.getElementById('menuToggle');
const profilePanel = document.getElementById('profilePanel');
const overlay = document.getElementById('overlay');
const closeProfileBtn = document.getElementById('closeProfileBtn');
const logoutFromProfile = document.getElementById('logoutFromProfile');

function openProfilePanel() {
    profilePanel.classList.add('open');
    overlay.classList.add('active');
}

function closeProfilePanel() {
    profilePanel.classList.remove('open');
    overlay.classList.remove('active');
}

if (menuToggle) {
    menuToggle.addEventListener('click', openProfilePanel);
}
if (closeProfileBtn) {
    closeProfileBtn.addEventListener('click', closeProfilePanel);
}
if (overlay) {
    overlay.addEventListener('click', closeProfilePanel);
}

// Выход из профиля
if (logoutFromProfile) {
    logoutFromProfile.addEventListener('click', function() {
        window.location.href = '/logout';
    });
}

function connectWebSocket() {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    ws = new WebSocket(protocol + "//" + window.location.host + "/ws");
    ws.onopen = () => console.log("WebSocket connected");
    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (currentChat && (data.from === currentChat || data.to === currentChat)) {
            appendMessage(data.from, data.text, data.from === currentUser ? "sent" : "received");
            // Прокрутка вниз
            const container = document.getElementById("messagesContainer");
            container.scrollTop = container.scrollHeight;
        }
        // Обновляем список диалогов, чтобы поднять актуальный чат наверх
        loadDialogs();
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
    document.getElementById("chatHeader").innerText = `${username}`;
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
    loadDialogs(); // обновить активный класс
}

function appendMessage(sender, text, direction) {
    const container = document.getElementById("messagesContainer");
    const div = document.createElement("div");
    div.className = `message ${direction}`;
    // Убираем блок с информацией об отправителе
    div.innerHTML = `<div class="message-bubble">${escapeHtml(text)}</div>`;
    container.appendChild(div);
    container.scrollTop = container.scrollHeight;
}

function sendMessage() {
    if (!currentChat) return;
    const input = document.getElementById("messageInput");
    const text = input.value.trim();
    if (!text) return;
    ws.send(JSON.stringify({ to: currentChat, text: text }));
    // Не добавляем сообщение сами, ждём подтверждения от сервера
    // Но можно сразу обновить список диалогов (чтобы поднять чат наверх)
    loadDialogs();
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

// Закрытие чата по ESC
function closeChat() {
    if (!currentChat) return;
    currentChat = null;
    document.getElementById("chatHeader").innerText = "";
    document.getElementById("inputArea").style.display = "none";
    document.getElementById("messagesContainer").innerHTML = "";
    document.querySelectorAll('.dialog-item').forEach(el => el.classList.remove('active'));
}

document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        e.preventDefault();
        // Если профильная панель открыта - закрываем её
        if (profilePanel && profilePanel.classList.contains('open')) {
            closeProfilePanel();
        } else {
            closeChat();
        }
        // Скрываем поисковые подсказки
        const results = document.getElementById("searchResults");
        if (results) results.style.display = "none";
    }
});

// Поиск пользователей
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

// === НОВЫЙ КОД: ресайзер левой панели ===
const sidebar = document.getElementById('sidebar');
const resizer = document.getElementById('resizer');

if (sidebar && resizer) {
    let startX, startWidth;

    function doResize(e) {
        const newWidth = startWidth + (e.clientX - startX);
        if (newWidth >= 180 && newWidth <= 500) {
            sidebar.style.width = newWidth + 'px';
            localStorage.setItem('sidebarWidth', newWidth);
        }
    }

    function stopResize() {
        document.documentElement.removeEventListener('mousemove', doResize);
        document.documentElement.removeEventListener('mouseup', stopResize);
    }

    resizer.addEventListener('mousedown', function(e) {
        startX = e.clientX;
        startWidth = parseInt(window.getComputedStyle(sidebar).width, 10);
        document.documentElement.addEventListener('mousemove', doResize);
        document.documentElement.addEventListener('mouseup', stopResize);
        e.preventDefault();
    });

    // Восстановить сохранённую ширину
    const saved = localStorage.getItem('sidebarWidth');
    if (saved) {
        const w = parseInt(saved, 10);
        if (w >= 180 && w <= 500) {
            sidebar.style.width = w + 'px';
        }
    }
}
// === КОНЕЦ НОВОГО КОДА ===

connectWebSocket();
loadDialogs();