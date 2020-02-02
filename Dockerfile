FROM golang:1.13.7-alpine3.11 AS build

# Copy source
WORKDIR /app/service-registry
COPY . .

# Download dependencies application
RUN go mod download

# Build application.
WORKDIR /app/service-registry/cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM alpine:3.11 AS run

WORKDIR /etc/service-registry/migrations
COPY ./resources/db/mysql/ .

WORKDIR /opt/app
RUN ls /etc/service-registry/migrations
COPY --from=build /app/service-registry/cmd/cmd service-registry
ENV GIN_MODE release
CMD ["./service-registry"]