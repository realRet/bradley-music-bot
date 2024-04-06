# syntax=docker/dockerfile:1
FROM golang:1.22

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application source code into the container
COPY . .

# Install ffmpeg
RUN apt-get update && apt-get install -y ffmpeg


# Build the Go binary
RUN go build -o music-bot

# Expose the port your Go application listens on
EXPOSE 8080

# Define the command to run when the container starts
CMD ["./music-bot"]


