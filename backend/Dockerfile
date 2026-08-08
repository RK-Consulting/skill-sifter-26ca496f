# Stage 1: Build
FROM golang:1.21-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 gives a statically linked binary that runs on a bare
# alpine base without needing matching libc/musl-dev at runtime
RUN CGO_ENABLED=0 GOOS=linux go build -o skillsifter .

# Stage 2: Minimal runtime image
FROM alpine:3.19

# ca-certificates needed for outbound TLS (e.g. sslmode=require to Postgres,
# any external HTTPS calls) — not present on a bare alpine image
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /app/skillsifter .

# Documentation only — with --network host (production) this has no effect;
# actual port is controlled by the PORT env var (see backend/.env)
EXPOSE 8080

CMD ["./skillsifter"]