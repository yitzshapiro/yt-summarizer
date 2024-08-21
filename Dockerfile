# Stage 1: Build the frontend
FROM node:18-alpine AS frontend-builder

# Set the working directory
WORKDIR /app

# Copy the frontend code
COPY ./yt-frontend .

# Install dependencies and build the frontend
RUN npm install && npm run build

# Stage 2: Build the backend and serve the frontend
FROM python:3.9-slim AS backend

# Install Go, Node.js, npm, and other necessary packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    golang-go \
    nodejs \
    npm \
    yt-dlp \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install http-server globally to serve the frontend
RUN npm install -g http-server

# Set the working directory for the Go backend
WORKDIR /go/src/app

# Copy go.mod and go.sum files first
COPY ./go.mod ./

# Download Go dependencies
RUN go mod download

# Copy the Go backend code
COPY ./main.go .

# Build the Go backend
RUN go build -o /app/main

# Set the working directory for Python scripts
WORKDIR /app

# Copy the Python scripts and requirements.txt
COPY ./summarize.py .
COPY ./transcribe.py .
COPY ./requirements.txt .

# Install Python dependencies, including torch
RUN python3 -m venv venv && \
    . ./venv/bin/activate && \
    pip install --upgrade pip && \
    pip install --no-cache-dir torch && \
    pip install --no-cache-dir -r requirements.txt

# Copy the frontend build from Stage 1
COPY --from=frontend-builder /app/dist /app/yt-frontend/dist

# Ensure the virtual environment is activated when running the Go application
ENV VIRTUAL_ENV=/app/venv
ENV PATH="$VIRTUAL_ENV/bin:$PATH"

# Start both the backend and the frontend servers
CMD ["sh", "-c", "./main & http-server /app/yt-frontend/dist -p 5137"]

# Expose the ports for both frontend and backend
EXPOSE 5001
EXPOSE 5137
