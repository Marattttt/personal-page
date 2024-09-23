FROM golang:1.23.1-alpine AS build

RUN apk add --update npm

WORKDIR /app/deps

RUN wget -O /app/deps/templ.tar.gz https://github.com/a-h/templ/releases/download/v0.2.778/templ_Linux_x86_64.tar.gz
RUN echo $PATH
RUN tar -xzf /app/deps/templ.tar.gz

ENV PATH="/app/deps:${PATH}"

WORKDIR /app/src

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

RUN go generate 
RUN go build -o /app/server ./cmd/frontend/

FROM golang:1.23.1-alpine AS release

WORKDIR /app

COPY --from=build /app/server /app/server

ENTRYPOINT ["./server"]

