# ---------- Build stage ----------
    FROM golang:1.22-alpine AS builder

    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    ARG SERVICE_PATH
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -o app ./${SERVICE_PATH}
    
    # ---------- Runtime stage ----------
    FROM alpine:3.19
    
    WORKDIR /app
    COPY --from=builder /app/app .
    
    EXPOSE 8080
    CMD ["./app"]