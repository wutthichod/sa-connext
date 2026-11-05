# Local MongoDB and NET-1 real-time features

The merged app uses the local `connext-mongo` image, the persistent MongoDB
volume from main, and NET-1's RabbitMQ, chat UI, and online-users page.

## Kubernetes with Tilt

Keep the existing ignored `infra/k8s/db-secrets.yaml` and
`infra/k8s/app-secrets.yaml`. Replace the `mongodb-secret` document in
`db-secrets.yaml` with `infra/k8s/mongodb-secret.example.yaml`, adjusting the
development credentials if needed. Set `RABBITMQ_URI` in `app-secret` to
`amqp://guest:guest@rabbitmq-svc:5672/`. Keep the existing JWT, Postgres, and
notification settings. All services must use the same broker and JWT secret.

Run `tilt up` against your local Kubernetes context. Tilt builds the MongoDB
image, waits for MongoDB and RabbitMQ, and exposes the gateway on port 8080.
`infra/k8s/kind-config.yaml` also maps gateway NodePort 30080 to host port 8080.
On Docker Desktop Kubernetes, NodePort 30080 can be used directly instead.

MongoDB initialization creates `chatuser` in `chat-db` on the first start with
an empty volume. The application authenticates with `authSource=chat-db`.
An existing initialized volume keeps its users and data; its credentials must
match the secret. Preserve that volume when restarting or upgrading.

The gateway deliberately uses one replica and the `Recreate` deployment
strategy. NET-1 stores presence and WebSocket connections in process memory,
and uses a shared `chat_gateway` queue. Increasing gateway replicas requires
shared presence and message routing to the gateway holding each connection.
Gateway restarts briefly disconnect clients, which reconnect automatically.

## Host services with Docker Compose infrastructure

Run `docker compose up -d --build` for infrastructure. Host-run Go services use:

```dotenv
RABBITMQ_URI=amqp://guest:guest@localhost:5672/
```

The chat service's ignored `services/chat-service/.env` also needs:

```dotenv
DATABASE_DSN=mongodb://chatuser:chatpassword@localhost:27017/chat-db?authSource=chat-db
CHAT_ADDR=:8082
```

Keep the user service's local Postgres settings and run its `setup` command
before `serve` on a fresh database. Setup migrates tables without dropping
existing user data.

## Frontend and verification

Run `npm install` then `npm run dev` in `_frontend`. Set
`NEXT_PUBLIC_BACKEND_URL=http://localhost:8080` in `_frontend/.env.local` and
restart Next.js after changing it. Both API proxy routes and WebSockets use
this setting. For a second device, use the server's reachable LAN IP and
gateway port instead of localhost, then open `http://<SERVER_IP>:3000`.

Log in with two different accounts in separate browser profiles. Both users
should appear on `/online-users`; joining and disconnecting should update
the other user's list immediately. Create a direct or group chat, send a
message, and check immediate receipt and history after refresh.

Run `go test ./...` and `npm run build` in `_frontend` for build checks.
The opt-in automated smoke test creates two users and chats, so run it against
a disposable local database:

```powershell
$env:CONNEXT_SMOKE_URL = 'http://localhost:8080'
go test ./demo -run TestRealtimeLocalStack -v -count=1
Remove-Item Env:CONNEXT_SMOKE_URL
```
