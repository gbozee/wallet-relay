# Stage 1: Build the Go application
FROM golang:bookworm AS builder

# Set the working directory within the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Set fixed environment variables for build
ENV DB_PATH="db"
ENV INDEX_PATH="templates/index.html"
ENV STATIC_PATH="templates/static"

# touch a .env
RUN touch .env

# Build the Go application
RUN go build -o main .

# Stage 2: Create a minimal image to run the Go application
FROM debian:bookworm-slim

# Set the working directory within the container
WORKDIR /app

# Copy the Go binary from the builder stage
COPY --from=builder /app/main /app/

# Copy any necessary files like templates, static assets, etc.
# COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/.env /app/

# Expose the port that the application will run on
EXPOSE 3355

# Set the command to run the executable
CMD ["./main"]