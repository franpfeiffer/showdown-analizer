let eventSource;
let reconnectAttempts = 0;
const maxReconnects = 10;
let lastRoomId = null;
let battleEnded = false;

const baseUrl = window.location.hostname === 'localhost'
    ? 'http://localhost:42069'
    : window.location.origin;

function resetUI() {
    const battleLog = document.getElementById('battle-log');
    const suggestionBox = document.getElementById('suggestion-box');
    const connectBtn = document.getElementById('connect-btn');
    const roomInput = document.getElementById('roomid-input');

    battleLog.innerHTML = '<p class="placeholder">Los eventos se muestran aca...</p>';
    suggestionBox.innerHTML = '';
    connectBtn.textContent = 'Conectar';
    connectBtn.disabled = false;
    roomInput.disabled = false;
    battleEnded = false;
    reconnectAttempts = 0;
}

function connectToBattle(roomid) {
    const battleLog = document.getElementById('battle-log');
    const suggestionBox = document.getElementById('suggestion-box');
    const connectBtn = document.getElementById('connect-btn');
    const roomInput = document.getElementById('roomid-input');

    battleLog.innerHTML = '<p class="placeholder">Conectando...</p>';
    suggestionBox.innerHTML = '';
    battleEnded = false;
    reconnectAttempts = 0;
    connectBtn.textContent = 'Conectando...';
    connectBtn.disabled = true;
    roomInput.disabled = true;

    if (eventSource) {
        eventSource.close();
    }

    eventSource = new EventSource(`${baseUrl}/connect?roomid=${encodeURIComponent(roomid)}`);
    lastRoomId = roomid;

    eventSource.onmessage = function(event) {
        if (battleLog.querySelector('.placeholder')) {
            battleLog.innerHTML = '';
        }

        if (event.data.includes('|win|') || event.data.includes('Batalla terminada')) {
            battleEnded = true;
            battleLog.innerHTML += '<p class="success">¡Batalla terminada! Puedes conectarte a otra batalla.</p>';
            connectBtn.textContent = 'Nueva Batalla';
            connectBtn.disabled = false;
            roomInput.disabled = false;
            roomInput.value = '';
            roomInput.focus();
        }

        if (event.data.includes("class='battle-summary'")) {
            suggestionBox.innerHTML = event.data;
        } else {
            const p = document.createElement('p');
            p.innerHTML = event.data;
            battleLog.appendChild(p);
            battleLog.scrollTop = battleLog.scrollHeight;
        }
    };

    eventSource.onerror = function(event) {
        if (battleEnded) {
            eventSource.close();
            return;
        }

        if (reconnectAttempts < maxReconnects) {
            reconnectAttempts++;
            battleLog.innerHTML += `<p class="warning">Reintentando conexión (${reconnectAttempts}/${maxReconnects})...</p>`;
            setTimeout(() => {
                if (!battleEnded) {
                    connectToBattle(lastRoomId);
                }
            }, 2000 * reconnectAttempts);
        } else {
            battleLog.innerHTML += '<p class="error">Error de conexión. Verifica el ID de la sala e intenta nuevamente.</p>';
            connectBtn.textContent = 'Reintentar';
            connectBtn.disabled = false;
            roomInput.disabled = false;
        }
        eventSource.close();
    };

    eventSource.onopen = function(event) {
        reconnectAttempts = 0;
        connectBtn.textContent = 'Conectado';
    };
}

document.getElementById('connect-form').addEventListener('submit', function(e) {
    e.preventDefault();
    const roomid = document.getElementById('roomid-input').value.trim();
    if (!roomid) {
        alert('Por favor ingresa un ID de sala válido');
        return;
    }
    connectToBattle(roomid);
});

document.getElementById('roomid-input').addEventListener('keypress', function(e) {
    if (e.key === 'Enter' && !this.disabled) {
        document.getElementById('connect-form').dispatchEvent(new Event('submit'));
    }
});
