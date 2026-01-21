# Chirpy API

Chirpy is a lightweight microblogging API. It provides user authentication, chirp creation, retrieval, deletion, and an admin interface for metrics and maintenance.

---

## Base URL

`http://<host>:<port>`

---

## Endpoints

### Health

`GET /api/healthz`  

**Response**

- `200 OK` — `OK`

---

## Users

### Create User

`POST /api/users`

**Request**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response**

`201 Created`

```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "<user@example.com>",
  "is_chirpy_red": false
}
```

### Update Email & Password

`PUT /api/users`
Requires authentication.

**Request**

```json

{
  "email": "<new@example.com>",
  "password": "newpassword"
}
  ```

**Response**

`200 OK`

```json
{
  "email": "<new@example.com>"
}
  ```

## Authentication

### Access Tokens (JWT)

Most user endpoints require a JWT access token in the `Authorization` header:

`Authorization: Bearer <token>`

### Refresh Tokens

Refresh tokens are also passed via `Authorization`:

`Authorization: Bearer <refresh_token>`

### Admin

Admin endpoints are protected by a server-side `IsAdmin` flag.

### Login

`POST /api/login`

**Request**

```json
{
  "email": "<user@example.com>",
  "password": "password123"
}
```

**Response**

`200 OK`

```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "<user@example.com>",
  "token": "<access_token>",
  "refresh_token": "<refresh_token>",
  "is_chirpy_red": false
}
  ```

#### Curl Example

```
curl -X POST /api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"<user@example.com>","password":"password123"}'
```

### Refresh Access Token

`POST /api/refresh`

Requires refresh token in Authorization.

**Response**

`200 OK`

```json
{
  "token": "<new_access_token>"
}
```

### Revoke Refresh Token

`POST /api/revoke`

Requires refresh token in Authorization.

**Response**

`204 No Content`

## Chirps

### Create Chirp

`POST /api/chirps`

Requires authentication. Chirps are limited to 140 characters and certain profanity is censored.

**Request**

```json
{
  "body": "Hello, Chirpy!"
}
  ```

**Response**

`201 Created`

```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "body": "Hello, Chirpy!",
  "user_id": "uuid"
}
  ```

#### Curl Example

```
curl -X POST /api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"body":"Hello, Chirpy!"}'
```

### Fetch Chirps

`GET /api/chirps`

**Query Parameters**

```
author_id (optional): UUID

sort (optional): asc (default) or desc
```

Response

`200 OK`

```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "body": "First chirp",
  "user_id": "uuid"
}
  ```

#### Curl Example

`curl /api/chirps?sort=desc`

### Fetch Chirp by ID

`GET /api/chirps/{chirpID}`

**Response**

```
200 OK (chirp object)

404 Not Found
```

### Delete Chirp

`DELETE /api/chirps/{chirpID}`

Requires authentication. Only the chirp owner may delete.

**Response**

```
204 No Content

403 Forbidden if not the owner
```

## Webhooks

### Polka Upgrade Webhook

`POST /api/polka/webhooks`

Requires an API key in the Authorization header.

**Request**

```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "uuid"
  }
}
  ```

**Response**

`204 No Content`

## Admin

### Metrics

`GET /admin/metrics`

Admin-only.

**Response**

`200 OK — HTML metrics page`

### Reset

`POST /admin/reset`

Admin-only. Deletes all users.

**Response**

`200 OK`

## Status Codes

Code | Meaning
---

200 | OK
201 | Created
204 | No Content
400 | Bad Request
401 | Unauthorized
403 | Forbidden
404 | Not Found
500 | Internal Server Error
