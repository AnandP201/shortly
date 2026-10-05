# ==============================================================================
# STAGE 1: Build the Go binary
# ==============================================================================
FROM golang:1.27-alpine AS builder

# Install git if your private dependencies require it
RUN apk add --no-cache git

RUN mkdir -p /code
ADD ./ /code
WORKDIR /code
COPY .env /code/.env

RUN go mod tidy -compat=1.17
# Build the application
# CGO_ENABLED=0 disables dynamic links for a fully self-contained static binary
# GOOS=linux ensures it is targeted to run in a Linux container
RUN CGO_ENABLED=0 GOOS=linux go build -o mainfile main.go

# ==============================================================================
# STAGE 2: Deployment runtime environment
# ==============================================================================
# Using 'alpine' keeps the image small (~10-20MB). Alternatively, use 'scratch' 
# if you want a completely empty base image (~server binary size only).
FROM alpine:3.20

RUN mkdir -p /code
WORKDIR /code

# Copy the pre-compiled binary from the builder stage
COPY --from=builder /code/.env /code/.env
COPY --from=builder /code/mainfile /code/mainfile

EXPOSE 8000

ENTRYPOINT  ["./mainfile"]
