FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Check if script exists in builder stage
RUN ls -al /app/script/

FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/pkg/persistence/sql/migrations ./migrations
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/script/run-migrations.sh .

# Convert Windows line endings to Unix line endings if needed
RUN sed -i 's/\r$//' run-migrations.sh

# Debugging: Verify the file is present and permissions are correct
RUN ls -al /root/

# Make script executable
RUN chmod +x run-migrations.sh

EXPOSE 8080

CMD ["/root/run-migrations.sh"]
