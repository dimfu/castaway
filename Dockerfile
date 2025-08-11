# Build Stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache nodejs npm make
RUN go install github.com/a-h/templ/cmd/templ@latest
WORKDIR /app
COPY package*.json ./
COPY Makefile .
COPY . .
RUN make build


# Run stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bin/main .
EXPOSE 6969
CMD ["./main", "-port", "6969", "-chunksize", "2"]
