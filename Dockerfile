# Step 1: Build Binary
FROM golang:1.17 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Accept the git revision as a build argument, defaulting to "development"
ARG GIT_REVISION=development

# Run tests and build with the git revision as a linker flag
RUN go test -v ./... && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-X main.gitRevision=$GIT_REVISION" -o main .

# Step 2: Create Executable Image
FROM debian:buster-slim

RUN apt-get update && \
    apt-get install -y ffmpeg && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /root/
COPY --from=builder /app/main .

ENTRYPOINT ["./main"]