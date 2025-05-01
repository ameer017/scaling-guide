FROM golang:1.23-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

ARG TARGET=api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app ./cmd/${TARGET}

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/app /app/server
EXPOSE 8080
ENTRYPOINT ["/app/server"]
