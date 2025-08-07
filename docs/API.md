# Turnate API Documentation

This document provides comprehensive documentation for the Turnate REST API.

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication

Most endpoints require authentication using JWT tokens. Include the token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

### Rate Limits
- **Global**: 10 requests/second, burst of 20
- **Auth endpoints**: 5 requests/minute  
- **API endpoints**: 5 requests/second, burst of 10

## Response Format

All responses are in JSON format:

### Success Response
```json
{
  "data": { ... },
  "message": "Success message"
}
```

### Error Response
```json
{
  "error": "Error type",
  "message": "Human readable error message",
  "field": "field_name" // (optional, for validation errors)
}
```

## Authentication Endpoints

### Register User
Create a new user account.

**Endpoint**: `POST /auth/register`

**Request Body**:
```json
{
  "username": "johndoe",
  "email": "john@example.com", 
  "password": "securepassword123",
  "display_name": "John Doe" // optional
}
```

**Response** (201 Created):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "01234567-89ab-7def-8901-234567890123",
    "username": "johndoe",
    "email": "john@example.com",
    "display_name": "John Doe",
    "role": "normal",
    "is_active": true
  },
  "message": "Registration successful! Welcome to Turnate! 🎉"
}
```

**Validation Rules**:
- `username`: 3-50 characters, alphanumeric and underscore only
- `email`: Valid email format
- `password`: Minimum 6 characters

### Login User
Authenticate user and receive JWT token.

**Endpoint**: `POST /auth/login`

**Request Body**:
```json
{
  "username": "johndoe", // or email
  "password": "securepassword123"
}
```

**Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "01234567-89ab-7def-8901-234567890123",
    "username": "johndoe", 
    "email": "john@example.com",
    "display_name": "John Doe",
    "role": "normal",
    "is_active": true
  },
  "message": "Login successful! Welcome back! 👋"
}
```

## User Endpoints

### Get Current User
Get the current user's profile.

**Endpoint**: `GET /users/me`
**Authentication**: Required

**Response** (200 OK):
```json
{
  "user": {
    "id": "01234567-89ab-7def-8901-234567890123",
    "username": "johndoe",
    "email": "john@example.com", 
    "display_name": "John Doe",
    "role": "normal",
    "is_active": true
  }
}
```

### List Users
Get a list of all users (basic info only).

**Endpoint**: `GET /users`
**Authentication**: Required

**Response** (200 OK):
```json
{
  "users": [
    {
      "id": "01234567-89ab-7def-8901-234567890123",
      "username": "johndoe",
      "display_name": "John Doe", 
      "role": "normal",
      "is_active": true
    }
  ]
}
```

### Update User
Update user profile. Users can update their own display_name. Admins can update role and is_active.

**Endpoint**: `PATCH /users/:id`
**Authentication**: Required

**Request Body**:
```json
{
  "display_name": "Updated Name", // anyone can update
  "role": "admin", // admin only
  "is_active": false // admin only
}
```

**Response** (200 OK):
```json
{
  "user": {
    "id": "01234567-89ab-7def-8901-234567890123",
    "username": "johndoe",
    "display_name": "Updated Name",
    "role": "normal", 
    "is_active": true
  },
  "message": "User updated successfully! ✅"
}
```

## Channel Endpoints

### List Channels
Get channels the user has access to.

**Endpoint**: `GET /channels`
**Authentication**: Required

**Response** (200 OK):
```json
{
  "channels": [
    {
      "id": "01234567-89ab-7def-8901-234567890124",
      "name": "general",
      "description": "General discussion",
      "type": "public",
      "created_by": "01234567-89ab-7def-8901-234567890123",
      "created_at": "2023-12-07T10:00:00Z", 
      "member_count": 5,
      "is_member": true
    }
  ]
}
```

### Create Channel
Create a new channel.

**Endpoint**: `POST /channels`
**Authentication**: Required

**Request Body**:
```json
{
  "name": "dev-team",
  "description": "Development team discussions", // optional
  "type": "public" // "public" or "private"
}
```

**Response** (201 Created):
```json
{
  "channel": {
    "id": "01234567-89ab-7def-8901-234567890125", 
    "name": "dev-team",
    "description": "Development team discussions",
    "type": "public",
    "created_by": "01234567-89ab-7def-8901-234567890123",
    "created_at": "2023-12-07T10:30:00Z",
    "member_count": 1,
    "is_member": true
  },
  "message": "Channel created successfully! 🎉"
}
```

**Notes**:
- Channel names are converted to lowercase with spaces replaced by hyphens
- Only admins can create private channels by default

### Get Channel Details
Get details about a specific channel.

**Endpoint**: `GET /channels/:id`
**Authentication**: Required

**Response** (200 OK):
```json
{
  "channel": {
    "id": "01234567-89ab-7def-8901-234567890124",
    "name": "general",
    "description": "General discussion",
    "type": "public", 
    "created_by": "01234567-89ab-7def-8901-234567890123",
    "created_at": "2023-12-07T10:00:00Z",
    "member_count": 5,
    "is_member": true
  }
}
```

### Join Channel
Join a public channel.

**Endpoint**: `POST /channels/:id/join`
**Authentication**: Required

**Response** (200 OK):
```json
{
  "message": "Successfully joined channel! 🎉"
}
```

**Notes**:
- Only public channels can be joined directly
- Admins can join any channel

### Leave Channel
Leave a channel.

**Endpoint**: `DELETE /channels/:id/leave` 
**Authentication**: Required

**Response** (200 OK):
```json
{
  "message": "Successfully left channel! 👋"
}
```

**Notes**:
- Cannot leave the "general" channel

### Get Channel Members
Get list of channel members.

**Endpoint**: `GET /channels/:id/members`
**Authentication**: Required

**Response** (200 OK):
```json
{
  "members": [
    {
      "id": "01234567-89ab-7def-8901-234567890123",
      "username": "johndoe",
      "display_name": "John Doe",
      "role": "normal",
      "is_active": true
    }
  ]
}
```

## Message Endpoints

### Send Message
Send a message to a channel.

**Endpoint**: `POST /channels/:channelId/messages`
**Authentication**: Required

**Request Body**:
```json
{
  "content": "Hello everyone! 👋",
  "thread_id": "01234567-89ab-7def-8901-234567890126" // optional, for replies
}
```

**Response** (201 Created):
```json
{
  "message": {
    "id": "01234567-89ab-7def-8901-234567890127",
    "content": "Hello everyone! 👋", 
    "user_id": "01234567-89ab-7def-8901-234567890123",
    "username": "johndoe",
    "display_name": "John Doe",
    "channel_id": "01234567-89ab-7def-8901-234567890124",
    "thread_id": null,
    "created_at": "2023-12-07T11:00:00Z",
    "updated_at": "2023-12-07T11:00:00Z",
    "reply_count": 0
  }
}
```

**Validation**:
- `content`: 1-2000 characters, required
- Must be a member of the channel

### Get Channel Messages
Get messages from a channel.

**Endpoint**: `GET /channels/:channelId/messages`
**Authentication**: Required

**Query Parameters**:
- `limit`: Number of messages (max 100, default 50)
- `offset`: Pagination offset (default 0)

**Response** (200 OK):
```json
{
  "messages": [
    {
      "id": "01234567-89ab-7def-8901-234567890127",
      "content": "Hello everyone! 👋",
      "user_id": "01234567-89ab-7def-8901-234567890123", 
      "username": "johndoe",
      "display_name": "John Doe",
      "channel_id": "01234567-89ab-7def-8901-234567890124",
      "thread_id": null,
      "created_at": "2023-12-07T11:00:00Z",
      "updated_at": "2023-12-07T11:00:00Z", 
      "reply_count": 3
    }
  ]
}
```

**Notes**:
- Returns only top-level messages (not thread replies)
- Messages ordered chronologically (oldest first)

### Get Thread Replies
Get replies to a threaded message.

**Endpoint**: `GET /channels/:channelId/messages/:threadId/replies`
**Authentication**: Required

**Query Parameters**:
- `limit`: Number of replies (max 100, default 50)
- `offset`: Pagination offset (default 0)

**Response** (200 OK):
```json
{
  "replies": [
    {
      "id": "01234567-89ab-7def-8901-234567890128",
      "content": "Great to see you here!",
      "user_id": "01234567-89ab-7def-8901-234567890123",
      "username": "janedoe", 
      "display_name": "Jane Doe",
      "channel_id": "01234567-89ab-7def-8901-234567890124",
      "thread_id": "01234567-89ab-7def-8901-234567890127",
      "created_at": "2023-12-07T11:05:00Z",
      "updated_at": "2023-12-07T11:05:00Z"
    }
  ]
}
```

### Get Recent Messages
Get recent messages across all user's channels.

**Endpoint**: `GET /messages/recent`
**Authentication**: Required

**Response** (200 OK):
```json
{
  "messages": [
    {
      "id": "01234567-89ab-7def-8901-234567890127",
      "content": "Hello everyone! 👋",
      "user_id": "01234567-89ab-7def-8901-234567890123",
      "username": "johndoe",
      "display_name": "John Doe", 
      "channel_id": "01234567-89ab-7def-8901-234567890124",
      "thread_id": null,
      "created_at": "2023-12-07T11:00:00Z",
      "updated_at": "2023-12-07T11:00:00Z",
      "reply_count": 3
    }
  ]
}
```

**Notes**:
- Returns messages from last 24 hours
- Limited to 20 most recent messages
- Only from channels user is a member of

## Admin Endpoints

### Get All Users (Admin)
Admin-only endpoint to get all users with full details.

**Endpoint**: `GET /admin/users`
**Authentication**: Required (Admin role)

**Response** (200 OK):
```json
{
  "users": [
    {
      "id": "01234567-89ab-7def-8901-234567890123",
      "username": "johndoe",
      "email": "john@example.com",
      "display_name": "John Doe",
      "role": "normal", 
      "is_active": true
    }
  ]
}
```

### Get All Channels (Admin)  
Admin-only endpoint to get all channels.

**Endpoint**: `GET /admin/channels`
**Authentication**: Required (Admin role)

**Response** (200 OK):
```json
{
  "channels": [
    {
      "id": "01234567-89ab-7def-8901-234567890124", 
      "name": "general",
      "description": "General discussion",
      "type": "public",
      "created_by": "01234567-89ab-7def-8901-234567890123",
      "created_at": "2023-12-07T10:00:00Z",
      "member_count": 5,
      "is_member": true
    }
  ]
}
```

## Error Codes

### HTTP Status Codes
- `200` - OK
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `415` - Unsupported Media Type
- `429` - Too Many Requests
- `500` - Internal Server Error

### Common Error Messages

#### Authentication Errors
```json
{
  "error": "Authorization header required"
}
```

```json
{
  "error": "Invalid token"
}
```

```json
{
  "error": "User not found or inactive"
}
```

#### Validation Errors
```json
{
  "error": "Invalid request data",
  "message": "Request contains potentially malicious content",
  "field": "username"
}
```

#### Rate Limiting
```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Please slow down."
}
```

#### Permission Errors
```json
{
  "error": "Admin access required"
}
```

```json
{
  "error": "Access denied to private channel"
}
```

## Data Types

### UUIDv7
All entity IDs use UUIDv7 format:
- Time-ordered UUIDs for better database performance
- Format: `01234567-89ab-7def-8901-234567890123`

### Timestamps
All timestamps are in ISO 8601 format with UTC timezone:
- Format: `2023-12-07T11:00:00Z`

### User Roles
- `normal` - Regular user with standard permissions
- `admin` - Administrator with elevated permissions

### Channel Types
- `public` - Anyone can join and see messages
- `private` - Invite-only, hidden from non-members

## SDK Examples

### JavaScript/Node.js
```javascript
const API_BASE = 'http://localhost:8080/api/v1';
const token = localStorage.getItem('jwt_token');

// Login
async function login(username, password) {
  const response = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  });
  
  const data = await response.json();
  if (response.ok) {
    localStorage.setItem('jwt_token', data.token);
    return data;
  }
  throw new Error(data.error);
}

// Send message
async function sendMessage(channelId, content) {
  const response = await fetch(`${API_BASE}/channels/${channelId}/messages`, {
    method: 'POST', 
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ content })
  });
  
  return response.json();
}
```

### curl Examples
```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Get channels (with token)
curl -X GET http://localhost:8080/api/v1/channels \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Send message
curl -X POST http://localhost:8080/api/v1/channels/CHANNEL_ID/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"content":"Hello from curl! 🚀"}'
```

## Best Practices

### Authentication
- Store JWT tokens securely (httpOnly cookies recommended for web apps)
- Implement automatic token refresh before expiration
- Handle 401 errors by redirecting to login

### Rate Limiting
- Implement exponential backoff for rate-limited requests
- Show user-friendly messages when rate limits are hit
- Cache data locally to reduce API calls

### Error Handling
- Always check HTTP status codes
- Display user-friendly error messages
- Log detailed errors for debugging

### Performance
- Use pagination for large data sets
- Implement client-side caching for frequently accessed data
- Batch multiple operations when possible

### Security
- Always use HTTPS in production
- Validate and sanitize all user inputs
- Never expose sensitive data in error messages