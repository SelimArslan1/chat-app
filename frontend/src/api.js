const API_BASE = '/api';

// Helper to get auth headers
function getAuthHeaders() {
    const token = localStorage.getItem('token');
    return token ? { Authorization: `Bearer ${token}` } : {};
}

// Generic fetch wrapper with error handling
async function apiRequest(endpoint, options = {}) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...getAuthHeaders(),
            ...options.headers,
        },
    });

    const data = await response.json().catch(() => ({}));

    if (!response.ok) {
        throw new Error(data.error || `Request failed with status ${response.status}`);
    }

    return data;
}

// ===== AUTH =====

export async function register(username, email, password) {
    const data = await apiRequest('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ username, email, password }),
    });
    if (data.access_token) {
        localStorage.setItem('token', data.access_token);
    }
    return data;
}

export async function login(email, password) {
    const data = await apiRequest('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    });
    if (data.access_token) {
        localStorage.setItem('token', data.access_token);
    }
    return data;
}

export async function getMe() {
    return apiRequest('/auth/me');
}

export function logout() {
    localStorage.removeItem('token');
}

export function getToken() {
    return localStorage.getItem('token');
}

// ===== SERVERS =====

export async function getServers() {
    return apiRequest('/servers');
}

export async function createServer(name) {
    return apiRequest('/servers', {
        method: 'POST',
        body: JSON.stringify({ name }),
    });
}

export async function getServer(serverId) {
    return apiRequest(`/servers/${serverId}`);
}

// ===== CHANNELS =====

export async function getChannels(serverId) {
    return apiRequest(`/servers/${serverId}/channels`);
}

export async function createChannel(serverId, name) {
    return apiRequest(`/servers/${serverId}/channels`, {
        method: 'POST',
        body: JSON.stringify({ name }),
    });
}

// ===== MESSAGES =====

export async function getMessages(channelId, limit = 50) {
    return apiRequest(`/channels/${channelId}/messages?limit=${limit}`);
}

export async function sendMessage(channelId, content) {
    return apiRequest(`/channels/${channelId}/messages`, {
        method: 'POST',
        body: JSON.stringify({ content }),
    });
}

export async function deleteMessage(messageId) {
    return apiRequest(`/messages/${messageId}`, {
        method: 'DELETE',
    });
}

// ===== WEBSOCKET =====

export function connectWebSocket(channelId, onMessage) {
    const token = getToken();
    if (!token) return null;

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/ws?token=${token}&channel_id=${channelId}`;

    const ws = new WebSocket(wsUrl);

    ws.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            onMessage(data);
        } catch (e) {
            console.error('WebSocket message parse error:', e);
        }
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };

    return ws;
}
