FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
WORKDIR /app/services/idp-mock
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN CGO_ENABLED=0 go build -o /app/idp-mock ./cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app/idp-mock .
EXPOSE 8001
CMD ["./idp-mock"]
