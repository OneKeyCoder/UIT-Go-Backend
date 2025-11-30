# Authentication Service

## Which database should use? Why?

We chose **PostgreSQL** for the Authentication Service because:
- User credentials require **strong consistency** and ACID transactions - a user's password change must be immediately reflected.
- The data model is relational with structured schema: `users` table with defined fields (id, email, password hash, etc.).
- PostgreSQL has mature security features for storing sensitive data.
- Team has experience with PostgreSQL, and it shares the same database infrastructure as Trip Service, reducing operational overhead.

We use **Azure Database for PostgreSQL - Flexible Server** with High Availability for production deployment.

## Why use JWT for authentication?

We chose **stateless JWT tokens** because:
- **Scalability**: No need for session storage or database lookups on every request.
- **Decoupling**: Any service can validate tokens independently without calling Auth Service.
- **Performance**: Token validation is just cryptographic verification, very fast.
- **Standard**: JWT is widely adopted, with good library support in Go.

We implement a **token pair strategy**:
- **Access Token** (24h expiry): Short-lived, used for API requests.
- **Refresh Token** (7d expiry): Long-lived, used only to get new access tokens.

This balances security (short access token lifetime) with user experience (don't need to re-login frequently).

## Why separate Authentication from Authorization?

Auth Service only handles **authentication** (who are you?), not **authorization** (what can you do?). Authorization is done by:
- Embedding user role in JWT claims.
- Each service checks JWT claims to authorize actions.

This keeps Auth Service simple and avoids it becoming a bottleneck for every request.
