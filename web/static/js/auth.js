const errorBox = document.getElementById('auth-error');

function showError(message) {
  if (errorBox) errorBox.textContent = message || '';
}

async function submitJSON(url, data) {
  const res = await fetch(url, {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    credentials: 'same-origin',
    body: JSON.stringify(data),
  });
  const payload = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(payload.error || 'Request failed');
  }
  return payload;
}

const loginForm = document.getElementById('login-form');
if (loginForm) {
  loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    showError('');
    const form = new FormData(loginForm);
    try {
      await submitJSON('/api/auth/login', {
        login: form.get('login'),
        password: form.get('password'),
      });
      window.location.href = '/chat';
    } catch (err) {
      showError(err.message);
    }
  });
}

const registerForm = document.getElementById('register-form');
if (registerForm) {
  registerForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    showError('');
    const form = new FormData(registerForm);
    try {
      await submitJSON('/api/auth/register', {
        username: form.get('username'),
        email: form.get('email'),
        password: form.get('password'),
      });
      window.location.href = '/chat';
    } catch (err) {
      showError(err.message);
    }
  });
}
