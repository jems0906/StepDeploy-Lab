FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
WORKDIR /app/services/protected-service
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app/protected-service ./cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app/protected-service .
EXPOSE 8004
CMD ["./protected-service"]
