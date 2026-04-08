FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY random.go service.go turnstile_vm.go ./
COPY cmd/render-api ./cmd/render-api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/render-api

FROM alpine:3.22

RUN adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=builder /out/app /app/app

ENV LISTEN_ADDR=0.0.0.0
ENV PORT=10000

EXPOSE 10000

USER appuser

ENTRYPOINT ["/app/app"]

