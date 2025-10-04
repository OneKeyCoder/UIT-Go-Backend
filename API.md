# API Documentation

## Base URLs

-   **Broker Service (API Gateway):** `http://localhost:8080`
-   **Authentication Service:** `http://localhost:8081`

---

## Authentication Service

### 1. Authenticate (Login)

Authenticates a user and returns JWT access and refresh tokens.

**Endpoint:** `POST /authenticate`

**Request Body:**

```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

**Validation Rules:**

-   `email`: Required, must be valid email format
-   `password`: Required, minimum 6 characters

**Success Response (200 OK):**

```json
{
    "error": false,
    "message": "Authentication successful",
    "data": {
        "user": {
            "id": 1,
            "email": "user@example.com",
            "first_name": "John",
            "last_name": "Doe",
            "active": 1,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        },
        "tokens": {
            "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
        }
    }
}
```

**Error Response (401 Unauthorized):**

```json
{
    "error": true,
    "message": "Invalid credentials"
}
```

**Validation Error (422 Unprocessable Entity):**

```json
{
    "error": true,
    "message": "Validation failed",
    "details": [
        {
            "field": "Email",
            "message": "Invalid email format",
            "code": "email"
        },
        {
            "field": "Password",
            "message": "Value is too short",
            "code": "min"
        }
    ]
}
```

**cURL Example:**

```bash
curl -X POST http://localhost:8081/authenticate \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

---

### 2. Refresh Token

Generates a new access token using a valid refresh token.

**Endpoint:** `POST /refresh`

**Request Body:**

```json
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Validation Rules:**

-   `refresh_token`: Required

**Success Response (200 OK):**

```json
{
    "error": false,
    "message": "Token refreshed successfully",
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
}
```

**Error Response (401 Unauthorized):**

```json
{
    "error": true,
    "message": "Invalid refresh token"
}
```

**cURL Example:**

```bash
curl -X POST http://localhost:8081/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

---

### 3. Validate Token

Validates a JWT access token (used by other services).

**Endpoint:** `POST /validate`

**Headers:**

```
Authorization: Bearer <access_token>
```

**Success Response (200 OK):**

```json
{
    "error": false,
    "message": "Token is valid",
    "data": {
        "user_id": 1,
        "email": "user@example.com",
        "role": "user",
        "exp": 1735689600,
        "iat": 1735603200,
        "nbf": 1735603200
    }
}
```

**Error Response (401 Unauthorized):**

```json
{
    "error": true,
    "message": "Invalid token"
}
```

**cURL Example:**

```bash
curl -X POST http://localhost:8081/validate \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

---

### 4. Health Check

Check if the authentication service is running.

**Endpoint:** `GET /ping`

**Success Response (200 OK):**

```
OK
```

**cURL Example:**

```bash
curl http://localhost:8081/ping
```

---

## Broker Service (API Gateway)

### 1. Broker Health Check

Check if the broker service is running.

**Endpoint:** `POST /`

**Success Response (200 OK):**

```json
{
    "error": false,
    "message": "Hit the broker",
    "data": null
}
```

**cURL Example:**

```bash
curl -X POST http://localhost:8080/
```

---

### 2. Handle Submission (Action Router)

Routes requests to appropriate microservices based on action type.

**Endpoint:** `POST /handle`

**Request Body:**

```json
{
  "action": "auth|log|mail",
  "auth": { ... },    // Required if action=auth
  "log": { ... },     // Required if action=log
  "mail": { ... }     // Required if action=mail
}
```

#### Action: Authentication (auth)

**Request:**

```json
{
    "action": "auth",
    "auth": {
        "email": "user@example.com",
        "password": "password123"
    }
}
```

**Response:** Same as Authentication Service `/authenticate` endpoint

**cURL Example:**

```bash
curl -X POST http://localhost:8080/handle \
  -H "Content-Type: application/json" \
  -d '{
    "action": "auth",
    "auth": {
      "email": "user@example.com",
      "password": "password123"
    }
  }'
```

#### Action: Logging (log)

**Request:**

```json
{
    "action": "log",
    "log": {
        "name": "broker-service",
        "data": "User logged in successfully"
    }
}
```

**Validation Rules:**

-   `log.name`: Required
-   `log.data`: Required

**Response:**

```json
{
    "error": false,
    "message": "Log entry created",
    "data": null
}
```

**cURL Example:**

```bash
curl -X POST http://localhost:8080/handle \
  -H "Content-Type: application/json" \
  -d '{
    "action": "log",
    "log": {
      "name": "test-service",
      "data": "Test log message"
    }
  }'
```

#### Action: Mail (mail)

**Request:**

```json
{
    "action": "mail",
    "mail": {
        "from": "sender@example.com",
        "to": "recipient@example.com",
        "subject": "Test Email",
        "message": "This is a test email message"
    }
}
```

**Validation Rules:**

-   `mail.from`: Required, valid email
-   `mail.to`: Required, valid email
-   `mail.subject`: Required
-   `mail.message`: Required

---

### 3. Log via gRPC

High-performance logging using gRPC protocol.

**Endpoint:** `POST /log-grpc`

**Request Body:**

```json
{
    "name": "service-name",
    "data": "Log message"
}
```

**Success Response (200 OK):**

```json
{
    "error": false,
    "message": "Log entry created via gRPC",
    "data": null
}
```

**cURL Example:**

```bash
curl -X POST http://localhost:8080/log-grpc \
  -H "Content-Type: application/json" \
  -d '{
    "name": "broker-service",
    "data": "Test gRPC log"
  }'
```

---

### 4. Health Check

Check if the broker service is running.

**Endpoint:** `GET /ping`

**Success Response (200 OK):**

```
OK
```

**cURL Example:**

```bash
curl http://localhost:8080/ping
```

---

## Common Response Formats

### Success Response

```json
{
  "error": false,
  "message": "Success message",
  "data": { ... }
}
```

### Error Response

```json
{
    "error": true,
    "message": "Error message"
}
```

### Validation Error Response

```json
{
    "error": true,
    "message": "Validation failed",
    "details": [
        {
            "field": "FieldName",
            "message": "Human-readable error message",
            "code": "validation_rule"
        }
    ]
}
```

---

## HTTP Status Codes

| Code | Meaning               | Usage                             |
| ---- | --------------------- | --------------------------------- |
| 200  | OK                    | Successful request                |
| 201  | Created               | Resource successfully created     |
| 400  | Bad Request           | Invalid request format            |
| 401  | Unauthorized          | Missing or invalid authentication |
| 403  | Forbidden             | Authenticated but not authorized  |
| 404  | Not Found             | Resource not found                |
| 422  | Unprocessable Entity  | Validation failed                 |
| 500  | Internal Server Error | Server error                      |

---

## JWT Token Structure

### Access Token (24h expiry)

```json
{
    "user_id": 1,
    "email": "user@example.com",
    "role": "user",
    "exp": 1735689600,
    "iat": 1735603200,
    "nbf": 1735603200
}
```

### Refresh Token (168h expiry)

```json
{
    "user_id": 1,
    "email": "user@example.com",
    "role": "",
    "exp": 1736208000,
    "iat": 1735603200,
    "nbf": 1735603200
}
```

**Note:** Refresh tokens don't include `role` for security reasons.

---

## Authentication Flow

```
1. Client → POST /authenticate
   ↓
2. Server validates credentials
   ↓
3. Server generates access + refresh tokens
   ↓
4. Client stores tokens
   ↓
5. Client includes access token in Authorization header:
   "Authorization: Bearer <access_token>"
   ↓
6. When access token expires:
   Client → POST /refresh with refresh_token
   ↓
7. Server validates refresh token
   ↓
8. Server returns new access + refresh tokens
```

---

## Rate Limiting (Future Implementation)

Recommended rate limits for production:

-   **Authentication:** 5 requests/minute per IP
-   **Token Refresh:** 10 requests/minute per user
-   **General API:** 100 requests/minute per user

---

## Postman Collection

You can import this API into Postman using the following structure:

```json
{
    "info": {
        "name": "UIT-Go Microservices",
        "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
    },
    "item": [
        {
            "name": "Authentication Service",
            "item": [
                {
                    "name": "Login",
                    "request": {
                        "method": "POST",
                        "url": "http://localhost:8081/authenticate",
                        "body": {
                            "mode": "raw",
                            "raw": "{\"email\":\"user@example.com\",\"password\":\"password123\"}"
                        }
                    }
                },
                {
                    "name": "Refresh Token",
                    "request": {
                        "method": "POST",
                        "url": "http://localhost:8081/refresh",
                        "body": {
                            "mode": "raw",
                            "raw": "{\"refresh_token\":\"{{refresh_token}}\"}"
                        }
                    }
                },
                {
                    "name": "Validate Token",
                    "request": {
                        "method": "POST",
                        "url": "http://localhost:8081/validate",
                        "header": [
                            {
                                "key": "Authorization",
                                "value": "Bearer {{access_token}}"
                            }
                        ]
                    }
                }
            ]
        }
    ]
}
```

---

## Testing with PowerShell

### Set Variables

```powershell
$baseUrl = "http://localhost:8081"
$email = "user@example.com"
$password = "password123"
```

### Login and Save Token

```powershell
$response = Invoke-RestMethod -Uri "$baseUrl/authenticate" `
  -Method Post `
  -ContentType "application/json" `
  -Body (@{email=$email; password=$password} | ConvertTo-Json)

$accessToken = $response.data.tokens.access_token
$refreshToken = $response.data.tokens.refresh_token
```

### Validate Token

```powershell
Invoke-RestMethod -Uri "$baseUrl/validate" `
  -Method Post `
  -Headers @{Authorization="Bearer $accessToken"}
```

### Refresh Token

```powershell
$newTokens = Invoke-RestMethod -Uri "$baseUrl/refresh" `
  -Method Post `
  -ContentType "application/json" `
  -Body (@{refresh_token=$refreshToken} | ConvertTo-Json)
```

---

For more information, see `MODERNIZATION.md` and `ARCHITECTURE.md`.
