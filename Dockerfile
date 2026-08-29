# syntax=docker/dockerfile:1

FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /urlshortener .

FROM gcr.io/distroless/base-debian11 AS final

COPY --from=builder /urlshortener /urlshortener

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/urlshortener"]