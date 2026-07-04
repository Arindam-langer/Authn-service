## Governance service

What is a governance service?
It is a service that determines whether to allow you access or not.
If you are a user then you are given a token else given a 401 or whatever is the status code.
Now what we can do is give him a role as well according to that and allow whatever it is we are providing.
Now in the power we can optimize it by using Redis and using different method of auths with different endpoints. there are different ways to role to a user i need to do research on that.


## the key things i want:
- It should handle both *Authorization*.
- It should be fast.
- It should be cool.

## The Important thing
the main thing is that i need to make this project be divided into small parts then i will be able to make it successfully.

### 1. HTTP Server & Routes setup (Completed)
- i have made the HTTP server and separated routing packages. it returns 200 response for health check.

### 2. Centralized Error Handling & Panic Recovery (Completed)
- i added a central `throwError` helper to handlers so that we log the actual error in the console but return clean JSON error messages to the client.
- i implemented `RecoveryMiddleware` so if anything panics, the server doesn't crash, it just recovers and throws a 500 error.

### 3. Database connection & Schema (Completed)
- using Postgres with docker-compose.
- updated schema to use `id SERIAL PRIMARY KEY`, `username UNIQUE`, `email UNIQUE`, and `phone_uuid UUID UNIQUE` along with password and timestamps.
- made a `UserStore` interface and `DBStore` struct to handle all queries and user creation.

### 4. Phone Number UUID v5 Authentication (Completed)
- client signs up using a phone number. the server deterministically generates a UUID v5 from it using standard `google/uuid` namespace hashing so we don't store raw phone numbers.
- when signing in, we lookup the user by their `phone_uuid`, verify they exist, and generate a JWT.
- the JWT stores the database `user_id` inside the token claims. The client carries this token, so the client never knows the actual database user ID.
- token validation is delegated to a separate package function `auth.VerifyToken()`.

### 5. Password Hashing (Completed)
- added `bcrypt` hashing on sign-up so we don't store raw passwords in the database.
- sign-in uses `bcrypt.CompareHashAndPassword` to verify the user password hash.

### 6. Code Refactoring (Completed)
- split `handlers.go` into three parts:
  - `types.go`: holds all request/response structs and a `Validate()` function on the struct itself (checks email format for `@` and domain characters, validates empty inputs).
  - `helpers.go`: holds `encode`, `decode`, and `throwError`.
  - `handlers.go`: only holds core endpoint logic, very clean.

### 7. Structured Logging with slog (Completed)
- switched from standard `log` to standard library `log/slog` which outputs structured JSON logs.

### 8. Config & Env Files (Completed)
- Removed hardcoded values for server configuration (address, timeouts), JWT secret, and Phone UUID namespace.
- Implemented environment variable loading with strict validation—failing fast at startup if configuration is missing or invalid.

### 9. Token Refresh Flow & Rotation (Completed)
- Designed a dedicated `refresh_tokens` table to store token hashes, expiration timestamps, and revocation status separate from the `users` table.
- Implemented secure Token Rotation (generating and storing a new refresh token and invalidating the old one on each refresh request).
- Built Token Reuse Detection (if a previously revoked refresh token is presented, all sessions for that user are immediately invalidated).
- Refactored token generation using a shared helper method `issueRefreshToken` to keep code clean and maintainable (DRY/KISS).

### 10. Session Revocation, Logout & Redis Blocklist (Completed)
- Added Redis to `docker-compose.yml` because we are cool like that and need speed.
- Implemented a `BlockStore` in Redis (`internal/db/redis.go`) to store blacklisted tokens.
- Added a `POST /signout` endpoint that invalidates the refresh token in Postgres and blocks the access token in Redis for its remaining lifetime (max 10 mins).
- Created a `BlocklistMiddleware` that intercepts incoming requests, hashes their access token, and checks Redis. If the token is revoked, it throws a `401 Unauthorized` ("token is revoked") right away.

### 11. Code Cleanup & Thread-Safety (Completed)
- Cleaned up `main.go` which was getting too fat. Moved environment variable parsing into `internal/config/config.go` so we fail-fast on startup.
- Upgraded Postgres storage from a single connection (`pgx.Conn`, which is NOT thread-safe and will crash if multiple users send requests simultaneously) to a proper connection pool (`pgxpool.Pool`) because we want this service to scale.
- Kept the project layout KISS & DRY—all server setup logic is back to being simple and sequential directly inside `func main()`.

**Initial Design Notes & Thoughts for refresh tokens & logout:**
> * **How to signOut?** What do we get in the request? We get an `accessToken` (in the Authorization header) and the `refreshToken` (in cookies). On logout, we need to invalidate/revoke both.
> * **Revoking the access token:** We can easily revoke the refresh token in the database, but since access tokens are stateless and not stored, how do we revoke the access token itself?
> * **Redis Blocklist:** We need to introduce Redis here to store revoked access tokens, along with a new middleware to check incoming `Authorization` headers. If the token is revoked, return `401 Unauthorized`.
> * **Two-Step Solution:**
>   1. Sign out endpoint revokes the refresh token and adds the access token to Redis.
>   2. Middleware checks if Redis has that token (using a simple lookup/hashset, similar to a Leetcode lookup).
>   3. Add redis in docker-compose because we are cool like that!!!!

---

## What is left to do:
1. *(Low priority)* **Database Migrations**: Setup a migration tool (like `goose` or `golang-migrate`) instead of blindly running `schema.sql` on startup.
2. *(Low priority)* **Redis Caching**: Store validated sessions or tokens in Redis to bypass database queries for active sessions.
3. **Observability**: Add Grafana and Prometheus because production needs eyes.
4. *(Low priority)* **Session Metadata Binding**: Bind each active refresh token to a specific device/browser session using client metadata (User-Agent, IP address) to prevent cookie hijacking and support active session lists.
