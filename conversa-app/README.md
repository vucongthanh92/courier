# Conversa

Conversa is a small React + TypeScript messaging client for the Courier chat-service MVP.

## Local setup

```bash
cd conversa-app
cp .env.example .env.local
npm install
npm run dev
```

The dev server uses `http://localhost:8080`, which matches the current local CORS settings in `user-service` and `chat-service`.

## Backend assumptions

- user-service: `http://localhost:5001/api/v1`
- chat-service: `http://localhost:5002/api/v1`

The app uses the real routes currently registered in backend `routes.go` files.
