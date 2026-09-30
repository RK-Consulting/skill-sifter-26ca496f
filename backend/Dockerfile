
FROM golang:1.21-alpine

ARG APP_VERSION=1.0.0
ARG VCS_REF=unknown

LABEL org.opencontainers.image.title="SkillSifter Backend" \
      org.opencontainers.image.version="$APP_VERSION" \
      org.opencontainers.image.revision="$VCS_REF" \
      org.opencontainers.image.description="SkillSifter v1.0.0 Go API"

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]

