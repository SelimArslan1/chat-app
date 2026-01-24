import { useEffect, useRef, useState, useCallback } from "react";
import * as api from "./api";

// ===== AUTH SCREEN COMPONENT =====
function AuthScreen({ onAuth }) {
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      if (isLogin) {
        await api.login(email, password);
      } else {
        await api.register(username, email, password);
      }
      const user = await api.getMe();
      onAuth(user);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-screen">
      <div className="auth-card">
        <h1>{isLogin ? "Welcome Back" : "Create Account"}</h1>
        <p>{isLogin ? "Sign in to continue chatting" : "Join the conversation"}</p>

        {error && <div className="error-message">{error}</div>}

        <form className="auth-form" onSubmit={handleSubmit}>
          {!isLogin && (
            <div className="form-group">
              <label>Username</label>
              <input
                type="text"
                placeholder="Enter username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                minLength={3}
              />
            </div>
          )}

          <div className="form-group">
            <label>Email</label>
            <input
              type="email"
              placeholder="Enter email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>

          <div className="form-group">
            <label>Password</label>
            <input
              type="password"
              placeholder="Enter password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={1}
            />
          </div>

          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? <span className="loading-spinner"></span> : (isLogin ? "Sign In" : "Create Account")}
          </button>
        </form>

        <div className="auth-toggle">
          {isLogin ? "Don't have an account? " : "Already have an account? "}
          <button className="btn btn-ghost" onClick={() => { setIsLogin(!isLogin); setError(""); }}>
            {isLogin ? "Sign Up" : "Sign In"}
          </button>
        </div>
      </div>
    </div>
  );
}

// ===== CREATE MODAL COMPONENT =====
function CreateModal({ title, placeholder, onSubmit, onClose }) {
  const [name, setName] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name.trim()) return;
    setLoading(true);
    try {
      await onSubmit(name.trim());
      onClose();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>{title}</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Name</label>
            <input
              type="text"
              placeholder={placeholder}
              value={name}
              onChange={(e) => setName(e.target.value)}
              autoFocus
              required
              minLength={1}
            />
          </div>
          <div className="modal-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? <span className="loading-spinner"></span> : "Create"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ===== INVITE MODAL COMPONENT =====
function InviteModal({ server, onClose, onInviteCreated }) {
  const [inviteCode, setInviteCode] = useState(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  const createInvite = async () => {
    setLoading(true);
    try {
      const invite = await api.createInvite(server.id);
      setInviteCode(invite.code);
      if (onInviteCreated) onInviteCreated(invite);
    } catch (err) {
      console.error("Failed to create invite:", err);
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(inviteCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>Invite to {server.name}</h2>
        <p>Share this code with friends to invite them to your server.</p>

        {inviteCode ? (
          <div className="invite-code-display">
            <input
              type="text"
              value={inviteCode}
              readOnly
              className="invite-code-input"
            />
            <button className="btn btn-primary" onClick={copyToClipboard}>
              {copied ? "Copied!" : "Copy"}
            </button>
          </div>
        ) : (
          <button
            className="btn btn-primary"
            onClick={createInvite}
            disabled={loading}
            style={{ width: "100%" }}
          >
            {loading ? <span className="loading-spinner"></span> : "Generate Invite Code"}
          </button>
        )}

        <div className="modal-actions">
          <button type="button" className="btn btn-secondary" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </div>
  );
}

// ===== JOIN SERVER MODAL COMPONENT =====
function JoinServerModal({ onClose, onJoined }) {
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!code.trim()) return;
    setError("");
    setLoading(true);

    try {
      const result = await api.joinServerWithCode(code.trim());
      if (onJoined) onJoined(result.server);
      onClose();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>Join a Server</h2>
        <p>Enter an invite code to join a server.</p>

        {error && <div className="error-message">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Invite Code</label>
            <input
              type="text"
              placeholder="Enter invite code"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              autoFocus
              required
            />
          </div>
          <div className="modal-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? <span className="loading-spinner"></span> : "Join Server"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ===== SERVER SIDEBAR COMPONENT =====
function ServerSidebar({ servers, selectedServer, onSelectServer, onCreateServer, onJoinServer }) {
  const [showModal, setShowModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);

  return (
    <>
      <div className="server-sidebar">
        {servers.filter(s => s && s.name).map((server) => (
          <div
            key={server.id}
            className={`server-icon ${selectedServer?.id === server.id ? "active" : ""}`}
            onClick={() => onSelectServer(server)}
            title={server.name || "Server"}
          >
            {(server.name || "S").charAt(0).toUpperCase()}
          </div>
        ))}

        <div className="server-divider"></div>

        <div
          className="server-icon add-server-btn"
          onClick={() => setShowModal(true)}
          title="Create Server"
        >
          +
        </div>

        <div
          className="server-icon join-server-btn"
          onClick={() => setShowJoinModal(true)}
          title="Join Server"
          style={{ background: "var(--bg-tertiary)", color: "var(--accent-secondary)" }}
        >
          →
        </div>
      </div>

      {showModal && (
        <CreateModal
          title="Create a Server"
          placeholder="My Awesome Server"
          onSubmit={onCreateServer}
          onClose={() => setShowModal(false)}
        />
      )}

      {showJoinModal && (
        <JoinServerModal
          onClose={() => setShowJoinModal(false)}
          onJoined={(server) => {
            if (onJoinServer) onJoinServer(server);
          }}
        />
      )}
    </>
  );
}

// ===== CHANNEL SIDEBAR COMPONENT =====
function ChannelSidebar({ server, channels, selectedChannel, onSelectChannel, onCreateChannel, user, onLogout }) {
  const [showModal, setShowModal] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);

  return (
    <>
      <div className="channel-sidebar">
        <div className="channel-header">
          <h2>{server?.name || "Select a Server"}</h2>
          <div style={{ display: "flex", gap: "4px" }}>
            {server && (
              <>
                <button
                  className="btn btn-ghost"
                  onClick={() => setShowInviteModal(true)}
                  title="Invite People"
                  style={{ fontSize: "0.9rem" }}
                >
                  📨
                </button>
                <button
                  className="btn btn-ghost"
                  onClick={() => setShowModal(true)}
                  title="Create Channel"
                >
                  +
                </button>
              </>
            )}
          </div>
        </div>

        <div className="channel-list">
          {server ? (
            <>
              <div className="channel-category">Text Channels</div>
              {channels.filter(c => c && c.id).map((channel) => (
                <div
                  key={channel.id}
                  className={`channel-item ${selectedChannel?.id === channel.id ? "active" : ""}`}
                  onClick={() => onSelectChannel(channel)}
                >
                  <span className="hash">#</span>
                  <span>{channel.name || "channel"}</span>
                </div>
              ))}
              {channels.length === 0 && (
                <div className="channel-item" style={{ opacity: 0.5, cursor: "default" }}>
                  No channels yet
                </div>
              )}
            </>
          ) : (
            <div className="empty-state">
              <p>Select or create a server to get started</p>
            </div>
          )}
        </div>

        <div className="user-panel">
          <div className="user-avatar">
            {user?.username?.charAt(0).toUpperCase() || "?"}
          </div>
          <div className="user-info">
            <div className="user-name">{user?.username || "User"}</div>
            <div className="user-status">Online</div>
          </div>
          <button className="logout-btn" onClick={onLogout} title="Logout">
            ⏻
          </button>
        </div>
      </div>

      {showModal && server && (
        <CreateModal
          title="Create a Channel"
          placeholder="general"
          onSubmit={(name) => onCreateChannel(server.id, name)}
          onClose={() => setShowModal(false)}
        />
      )}

      {showInviteModal && server && (
        <InviteModal
          server={server}
          onClose={() => setShowInviteModal(false)}
        />
      )}
    </>
  );
}

// ===== MESSAGE ITEM COMPONENT =====
function MessageItem({ message }) {
  const formatTime = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  };

  return (
    <div className="message-item">
      <div className="message-avatar">
        {(message.username || "?").charAt(0).toUpperCase()}
      </div>
      <div className="message-content">
        <div className="message-header">
          <span className="message-author">{message.username || "Unknown User"}</span>
          <span className="message-time">{formatTime(message.created_at)}</span>
        </div>
        <div className="message-text">{message.content}</div>
      </div>
    </div>
  );
}

// ===== CHAT AREA COMPONENT =====
function ChatArea({ channel, messages, onSendMessage }) {
  const [input, setInput] = useState("");
  const messagesEndRef = useRef(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = () => {
    if (!input.trim()) return;
    onSendMessage(input.trim());
    setInput("");
  };

  const handleKeyDown = (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  if (!channel) {
    return (
      <div className="chat-area">
        <div className="empty-state">
          <div className="empty-state-icon">💬</div>
          <h3>No Channel Selected</h3>
          <p>Select a channel from the sidebar to start chatting</p>
        </div>
      </div>
    );
  }

  return (
    <div className="chat-area">
      <div className="chat-header">
        <span className="hash">#</span>
        <h3>{channel.name}</h3>
      </div>

      <div className="messages-container">
        <div className="messages-list">
          {messages.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state-icon">🎉</div>
              <h3>Welcome to #{channel.name}</h3>
              <p>This is the beginning of the channel. Send a message to get started!</p>
            </div>
          ) : (
            messages.map((msg) => <MessageItem key={msg.id} message={msg} />)
          )}
          <div ref={messagesEndRef} />
        </div>
      </div>

      <div className="message-input-container">
        <div className="message-input-wrapper">
          <input
            className="message-input"
            type="text"
            placeholder={`Message #${channel.name}`}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
          />
          <button className="send-btn" onClick={handleSend}>
            ➤
          </button>
        </div>
      </div>
    </div>
  );
}

// ===== MAIN APP COMPONENT =====
export default function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [servers, setServers] = useState([]);
  const [selectedServer, setSelectedServer] = useState(null);
  const [channels, setChannels] = useState([]);
  const [selectedChannel, setSelectedChannel] = useState(null);
  const [messages, setMessages] = useState([]);
  const wsRef = useRef(null);

  // Check for existing token on mount
  useEffect(() => {
    const checkAuth = async () => {
      const token = api.getToken();
      if (token) {
        try {
          const userData = await api.getMe();
          setUser(userData);
        } catch {
          api.logout();
        }
      }
      setLoading(false);
    };
    checkAuth();
  }, []);

  // Load servers when user is authenticated
  useEffect(() => {
    if (user) {
      loadServers();
    }
  }, [user]);

  // Load channels when server is selected
  useEffect(() => {
    if (selectedServer) {
      loadChannels(selectedServer.id);
    } else {
      setChannels([]);
      setSelectedChannel(null);
    }
  }, [selectedServer]);

  // Connect WebSocket and load messages when channel is selected
  useEffect(() => {
    if (selectedChannel) {
      loadMessages(selectedChannel.id);
      connectToChannel(selectedChannel.id);
    } else {
      setMessages([]);
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    }

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [selectedChannel]);

  const loadServers = async () => {
    try {
      const data = await api.getServers();
      setServers(data || []);
    } catch (err) {
      console.error("Failed to load servers:", err);
    }
  };

  const loadChannels = async (serverId) => {
    try {
      const data = await api.getChannels(serverId);
      setChannels(data || []);
    } catch (err) {
      console.error("Failed to load channels:", err);
    }
  };

  const loadMessages = async (channelId) => {
    try {
      const data = await api.getMessages(channelId);
      // Messages come in DESC order, reverse for display
      setMessages((data || []).reverse());
    } catch (err) {
      console.error("Failed to load messages:", err);
    }
  };

  const connectToChannel = useCallback((channelId) => {
    if (wsRef.current) {
      wsRef.current.close();
    }

    const ws = api.connectWebSocket(channelId, (event) => {
      if (event.type === "NEW_MESSAGE") {
        setMessages((prev) => [...prev, event.payload]);
      }
    });

    wsRef.current = ws;
  }, []);

  const handleAuth = (userData) => {
    setUser(userData);
  };

  const handleLogout = () => {
    api.logout();
    setUser(null);
    setServers([]);
    setSelectedServer(null);
    setChannels([]);
    setSelectedChannel(null);
    setMessages([]);
  };

  const handleCreateServer = async (name) => {
    try {
      const newServer = await api.createServer(name);
      console.log("Created server:", newServer);
      if (newServer && newServer.id) {
        setServers((prev) => [...prev, newServer]);
        setSelectedServer(newServer);
      } else {
        console.error("Server created but response missing id:", newServer);
        // Reload servers from API to ensure we have latest data
        await loadServers();
      }
    } catch (err) {
      console.error("Failed to create server:", err);
    }
  };

  const handleSelectServer = (server) => {
    setSelectedServer(server);
    setSelectedChannel(null);
  };

  const handleCreateChannel = async (serverId, name) => {
    const newChannel = await api.createChannel(serverId, name);
    setChannels((prev) => [...prev, newChannel]);
    setSelectedChannel(newChannel);
  };

  const handleSelectChannel = (channel) => {
    setSelectedChannel(channel);
  };

  const handleSendMessage = (content) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: "SEND_MESSAGE",
        content,
      }));
    }
  };

  const handleJoinServer = (server) => {
    if (server && server.id) {
      setServers((prev) => [...prev, server]);
      setSelectedServer(server);
    } else {
      // Reload servers to get the joined one
      loadServers();
    }
  };

  if (loading) {
    return (
      <div className="auth-screen">
        <div className="loading-spinner" style={{ width: 40, height: 40 }}></div>
      </div>
    );
  }

  if (!user) {
    return <AuthScreen onAuth={handleAuth} />;
  }

  return (
    <div className="app-container">
      <ServerSidebar
        servers={servers}
        selectedServer={selectedServer}
        onSelectServer={handleSelectServer}
        onCreateServer={handleCreateServer}
        onJoinServer={handleJoinServer}
      />

      <ChannelSidebar
        server={selectedServer}
        channels={channels}
        selectedChannel={selectedChannel}
        onSelectChannel={handleSelectChannel}
        onCreateChannel={handleCreateChannel}
        user={user}
        onLogout={handleLogout}
      />

      <ChatArea
        channel={selectedChannel}
        messages={messages}
        onSendMessage={handleSendMessage}
      />
    </div>
  );
}
