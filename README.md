# Universal AI Prompt Service

A global AI prompt microservice built with Go (Gin), supporting REST, gRPC, and WebSocket communication. This service acts as a centralized, reusable prompt and response manager with intelligent caching and real-time communication capabilities.

## Features

- **Universal System Prompt Store**  
  Store and manage multiple system prompts for different tenants or contexts.

- **Prompt Request Caching with Hashing**  
  Avoid duplicate AI requests by generating an MD5 hash for identical input and reusing cached responses from the database.

- **AI Response Proxying**  
  Seamlessly integrates with OpenAI or similar APIs to forward prompts and retrieve completions.

- **WebSocket Chat Support**  
  Real-time bidirectional WebSocket communication with per-user rooms for live chat experiences.

- **REST & gRPC Support**  
  Expose endpoints for both REST and gRPC protocols to support any type of client.

- **Multi-Tenant Ready**  
  Designed with tenant-level prompt isolation for SaaS environments.

## Database Schema

- `prompts` — Stores reusable system prompts.
- `prompt_cache` — Caches hashed prompt+input requests and AI responses.
- `users` — Tracks users for WebSocket room management.

## Example Use Cases

- AI chatbots
- Prompt-as-a-service for internal teams
- Multi-tenant SaaS AI integrations
- Real-time WebSocket chat apps with AI

## Tech Stack

- Go (Gin Framework)
- WebSockets
- PostgreSQL
- gRPC
- REST (JSON API)
- Optional: Redis for faster cache lookup

## Future Additions

- Admin dashboard for prompt management  
- TTL-based cache cleanup  
- Rate limiting and auth middlewares  
