FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o /bin/agent-gateway .

FROM alpine:3.22

WORKDIR /app
COPY --from=builder /bin/agent-gateway /bin/agent-gateway
COPY config ./config

ENV HTTP_PORT=:5010
EXPOSE 5010

ENTRYPOINT ["/bin/agent-gateway"]
CMD ["--config=./config/local/config.yaml"]
