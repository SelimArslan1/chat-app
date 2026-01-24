# Real-Time Chat Application

A production-ready real-time chat platform demonstrating modern backend engineering practices with Go, WebSockets, and PostgreSQL.

### Authentication & Security
- **JWT-based auth** with short-lived access tokens (15 min) and long-lived refresh tokens (7 days)
- **Token revocation** via database-stored refresh tokens - prevents token reuse after logout
- **Secure password hashing** with bcrypt
- **Rate limiting middleware** - IP-based and user-based limits to prevent abuse
- **CORS configuration** for WebSocket connections

### Real-Time Communication
- **WebSocket implementation** using gorilla/websocket with proper connection lifecycle
- **Hub pattern** for managing concurrent connections with mutex-protected client registry
- **Per-user message rate limiting** (30 msg/min) directly in WebSocket handler
- **Graceful error handling** - rate limit errors sent as WebSocket events

## Tech Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.24 |
| **Framework** | Gin |
| **ORM** | GORM |
| **Database** | PostgreSQL 16 |
| **WebSocket** | gorilla/websocket |
| **Auth** | JWT (golang-jwt/jwt/v5) |
| **Password** | bcrypt |
| **Frontend** | React + Vite |

## Security Implementation

### Rate Limiting Strategy
```go
// Three-tier rate limiting:
// 1. Global: 60 req/min per IP (all endpoints)
// 2. Strict: 5 req/min per IP (auth endpoints - prevents brute force)
// 3. Message: 30 msg/min per user (WebSocket + HTTP)
```

## WebSocket Architecture

- **Hub** maintains map of clients, protected by `sync.RWMutex`
- **Broadcast** sends to all clients in same channel
- **Non-blocking sends** prevent slow clients from blocking others

## Database Schema

7 tables with proper relationships:
- **Users** - Authentication, profile
- **Servers** - Discord-like server entities
- **ServerMembers** - Many-to-many with roles
- **Channels** - Belong to servers
- **Messages** - Soft-deletable, with username denormalization for display
- **ServerInvites** - Expirable, usage-limited invite codes
- **RefreshTokens** - Revocable token storage

## Running the Project

```bash
docker-compose up --build
# Frontend: http://localhost:3000
# API: http://localhost:8080
```

