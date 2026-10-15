# Stage 1: Build stage using official Go Linux image
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first (optimizes Docker layer caching)
COPY go.mod ./

# 1. Explicitly download and cache dependencies inside the container
#RUN go mod download

# Copy source code and HTML templates
COPY . .

# Build a static Linux binary
#RUN CGO_ENABLED=0 GOOS=linux go build -o weather-app main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o weather-app .

# Stage 2: Lightweight runtime stage
FROM alpine:latest

WORKDIR /root/

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/weather-app .

# Copy the index.html file needed for template rendering
COPY --from=builder /app/index.html .

# Expose port 8080
EXPOSE 8080

# Command to execute when the container starts
CMD ["./weather-app"]


