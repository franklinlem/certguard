FROM golang:1.27.3-alpine AS build
WORKDIR /src
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /certguard ./cmd/certguard
FROM alpine:3.24
RUN apk add --no-cache ca-certificates && addgroup -g 10001 app && adduser -D -u 10001 -G app app
COPY --from=build /certguard /usr/local/bin/certguard
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["certguard"]
CMD ["serve"]
