FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/go-api .
FROM alpine:3.21
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /out/go-api /app/go-api
EXPOSE 8080
USER 10001
ENTRYPOINT ["/app/go-api"]
