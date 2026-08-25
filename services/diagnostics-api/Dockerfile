FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
WORKDIR /app/services/diagnostics-api
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app/diagnostics-api ./cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app/diagnostics-api .
EXPOSE 8080
CMD ["./diagnostics-api"]
