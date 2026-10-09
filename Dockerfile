FROM golang:1.27.2-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go run ./cmd/checkstructure && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /harkai-api ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && adduser -D -H app
USER app
COPY --from=build /harkai-api /usr/local/bin/harkai-api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/harkai-api"]
