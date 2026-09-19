FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY pkg/ ./pkg/

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/bin/server ./cmd

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/bin/server /server

EXPOSE 3000

ENTRYPOINT ["/server"]
