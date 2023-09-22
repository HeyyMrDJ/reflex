# Use the official Golang image as the base image
FROM golang:1.21.1-alpine3.18 AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code and template files into the container
COPY main.go /app/
COPY go.mod /app/
COPY go.sum /app/
COPY templates/ /app/templates/
COPY static/ /app/static/

# Install Go dependencies using Go modules
RUN go mod download

# Build the Go application
RUN go build -o reflex main.go

# Start a new stage for the lightweight Alpine image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the previous stage
COPY --from=build /app/reflex /app/reflex
COPY --from=build /app/templates /app/templates
COPY --from=build /app/static /app/static

# Expose the port that the Go application will run on
EXPOSE 9069

# Run the Go application when the container starts
CMD ["./reflex"]
