# Stage 1: Build Frontend
FROM node:18 AS frontend-builder

WORKDIR /app/frontend

COPY ./yt-frontend/package.json ./yt-frontend/pnpm-lock.yaml ./
COPY ./yt-frontend ./

RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build

# Stage 2: Build Backend
FROM python:3.9-slim AS backend-builder

WORKDIR /app

# Copy Go files for the backend
COPY ./main.go ./go.mod ./

# Install Go and build the Go application
RUN apt-get update && apt-get install -y golang
RUN CGO_ENABLED=0 go build -o /app/server

# Install system-wide dependencies
RUN apt-get update && apt-get install -y ffmpeg curl

# Install Python dependencies system-wide, including Whisper
RUN pip install --upgrade pip
RUN pip install openai-whisper yt-dlp

# Copy the Python scripts
COPY ./summarize.py ./transcribe.py ./

# Stage 3: Production Environment
FROM python:3.9-slim

WORKDIR /app

# Install necessary packages system-wide
RUN apt-get update && apt-get install -y ffmpeg curl

# Install Ollama and pull the model
RUN curl -fsSL https://ollama.com/install.sh | sh

# Copy the backend server and Python scripts
COPY --from=backend-builder /app/server /app/server
COPY --from=backend-builder /app/summarize.py /app/summarize.py
COPY --from=backend-builder /app/transcribe.py /app/transcribe.py

# Copy the frontend build
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Expose ports for frontend and backend
EXPOSE 5001
EXPOSE 8080

# Start all services
CMD ["sh", "-c", "ollama serve & sleep 5 && ollama pull qwen2:1.5b && ./server & python3 -m http.server --directory /app/frontend/dist 8080"]
