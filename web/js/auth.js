const byId = (id) => document.getElementById(id);

document.addEventListener("DOMContentLoaded", () => {
    byId("registerBtn").addEventListener("click", register);
    byId("loginBtn").addEventListener("click", login);
});

function setStatus(text) {
    const el = byId("status");
    if (el) el.textContent = text;
}

async function register() {
    setStatus("Registering...");

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
    alert("Registered. Now login.");
}

async function login() {
    setStatus("Logging in...");

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
    localStorage.setItem("token", data.token);

    window.location.href = "/chat.html";
}