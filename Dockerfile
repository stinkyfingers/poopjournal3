FROM golang:1.26.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o /out/poopjournal .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /out/poopjournal /app/poopjournal

ENV PORT=8080
ENV STORAGE_TYPE=s3
ENV S3_BUCKET=poopjournal-data-prod

EXPOSE 8080

CMD ["/app/poopjournal"]