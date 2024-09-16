FROM golang:1.23.1-alpine AS build

WORKDIR /app/deps

# Install templ
RUN wget -O /app/deps/templ.tar.gz https://github.com/a-h/templ/releases/download/v0.2.778/templ_Linux_x86_64.tar.gz && \
	 tar -xzf /app/deps/templ.tar.gz

# Include templ in PATH
ENV PATH="/app/deps:${PATH}"

WORKDIR /app/src

# Install dependencies (for cache)
COPY go.mod go.sum ./
RUN go mod download 

# Build 
COPY . .
RUN go build -o /app/server ./cmd/mq/

# Add base scripts
FROM marattttt/runnerbase AS runnerbase

FROM golang:1.23.1-alpine AS release

WORKDIR /app

RUN apk update && apk add bash sudo

# Create a new user
COPY --from=runnerbase /scripts/create_user_busybox.sh /app/scripts/create_user.sh
RUN sh /app/scripts/create_user.sh

COPY --from=build /app/server /app/server

ENTRYPOINT ["/app/server"]
