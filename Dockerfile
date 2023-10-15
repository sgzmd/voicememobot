# Step 1: Build Binary
FROM golang:1.21 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Accept the git revision as a build argument, defaulting to "development"
ARG GIT_REVISION=development

# Run tests and build with the git revision as a linker flag
RUN go test -v ./... && \
    go build -ldflags="-X main.gitRevision=$GIT_REVISION" -o main cmd/main.go

# Step 2: Create Executable Image
FROM alpine:3.18.4

# Install ffmpeg and ca-certificates
RUN apk --no-cache add ffmpeg ca-certificates

WORKDIR /root/
COPY --from=builder /app/main .

ENTRYPOINT ["./main"]