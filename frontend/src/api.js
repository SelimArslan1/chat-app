const API_BASE = '/api';

// Helper to get auth headers
function getAuthHeaders() {
    const token = localStorage.getItem('token');
    return token ? { Authorization: `Bearer ${token}` } : {};
}

// Generic fetch wrapper with error handling and auto-refresh
async function apiRequest(endpoint, options = {}, retry = true) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...getAuthHeaders(),
            ...options.headers,
        },
    });

    const data = await response.json().catch(() => ({}));

    // Try to refresh token on 401 and retry once
    if (response.status === 401 && retry) {
        const refreshed = await tryRefreshToken();
        if (refreshed) {
            return apiRequest(endpoint, options, false);
        }
    }

    // Handle rate limit errors with user-friendly message
    if (response.status === 429) {
        const retryAfter = data.retry_after || 60;
        throw new Error(`Too many requests. Please wait ${Math.ceil(retryAfter)} seconds and try again.`);
    }

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
    }, false);
    if (data.access_token) {
        localStorage.setItem('token', data.access_token);
    }
    if (data.refresh_token) {
        localStorage.setItem('refresh_token', data.refresh_token);
    }
    return data;
}

export async function login(email, password) {
    const data = await apiRequest('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    }, false);
    if (data.access_token) {
        localStorage.setItem('token', data.access_token);
    }
    if (data.refresh_token) {
        localStorage.setItem('refresh_token', data.refresh_token);
    }
    return data;
}

export async function tryRefreshToken() {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) return false;

    try {
        const response = await fetch(`${API_BASE}/auth/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken }),
        });

        if (!response.ok) {
            logout();
            return false;
        }

        const data = await response.json();
        if (data.access_token) {
            localStorage.setItem('token', data.access_token);
        }
        if (data.refresh_token) {
            localStorage.setItem('refresh_token', data.refresh_token);
        }
        return true;
    } catch {
        logout();
        return false;
    }
}

export async function getMe() {
    return apiRequest('/auth/me');
}

export async function logout() {
    const refreshToken = localStorage.getItem('refresh_token');

    // Revoke token on server
    if (refreshToken) {
        try {
            await fetch(`${API_BASE}/auth/logout`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ refresh_token: refreshToken }),
            });
        } catch {
            // Ignore errors, still clear local storage
        }
    }

    localStorage.removeItem('token');
    localStorage.removeItem('refresh_token');
}

export async function logoutAll() {
    await apiRequest('/auth/logout-all', { method: 'POST' });
    localStorage.removeItem('token');
    localStorage.removeItem('refresh_token');
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

// ===== INVITES =====

export async function createInvite(serverId, maxUses = 0, expiresIn = 0) {
    return apiRequest(`/servers/${serverId}/invites`, {
        method: 'POST',
        body: JSON.stringify({ max_uses: maxUses, expires_in: expiresIn }),
    });
}

export async function getInvites(serverId) {
    return apiRequest(`/servers/${serverId}/invites`);
}

export async function deleteInvite(serverId, inviteId) {
    return apiRequest(`/servers/${serverId}/invites/${inviteId}`, {
        method: 'DELETE',
    });
}

export async function joinServerWithCode(code) {
    return apiRequest('/invites/join', {
        method: 'POST',
        body: JSON.stringify({ code }),
    });
}

// ===== WEBSOCKET =====

export function connectWebSocket(channelId, onMessage, onError) {
    const token = getToken();
    if (!token) return null;

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/ws?token=${token}&channel_id=${channelId}`;

    const ws = new WebSocket(wsUrl);

    ws.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);

            if (data.type === 'ERROR' && onError) {
                onError(data.payload);
                return;
            }

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
