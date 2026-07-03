FROM golang:1.26-alpine AS build

RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service .

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /app/user-service /app/user-service
COPY --from=build /app/config /app/config
COPY --from=build /app/docs /app/docs
COPY --from=build /app/internal/resources /app/internal/resources

EXPOSE 5000 5001

USER nonroot:nonroot
ENTRYPOINT ["/app/user-service"]
CMD ["start"]
