## Governance service

What is a governance service?
It is a service that determines whether to allow you access or not.
If you are a user then you are given a token else given a 401 or whatever is the status code.
Now what we can do is give him a role as well according to that and allow whatever it is we are providing.
Now in the power we can optimize it by using Redis and using different method of auths with different endpoints. there are different ways to role to a user i need to do research on that.


## the key things i want:
- It should handle both *Authentication* and *Authorization*.
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

---

## What is left to do:
1. *(Low priority)* **Database Migrations**: Setup a migration tool (like `goose` or `golang-migrate`) instead of blindly running `schema.sql` on startup.
2. **Token Refresh Flow**: Add a refresh token mechanism stored in HTTP-only cookies so the user doesn't get logged out after 10 minutes.
    - i mean in this sense we create a sign up request do we create a refresh token then. or when user signs up and checks if our refresh token is expired then generate one more now, i think when we sign in only then we switch or create token. now how does sign in work sign in gives a refresh token as well as access token.
    we get a sign in request we verify user and then provide access token and new refresh token if expired as old on in cookies 
    now in sign in we do two things we make a refresh token and access token everytime because checking is time consuming rather than just make a new token instead
    but the access token goes in cookies to prevent xss attack by attacker using js apparently. now we can use the same token generate function but with different password should be different. now in sign up function we need to add just three things one to create a refresh token, pass it to cookies using net/http and store it in db and update it every time
3. **Redis Caching**: Store validated sessions or tokens in Redis to bypass database queries for active sessions
4. Add Grafana and Promethius


we need to add revocation when token is expired and send 401 or something when we get and token mismatch as well we need to take care of that when we get 
