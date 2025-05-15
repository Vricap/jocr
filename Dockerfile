# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

# Copy your source code
COPY . .

# Build the Go binary
RUN go build -o main .

# Runtime stage
FROM debian\:bookworm-slim

# Install Tesseract and the Javanese language pack
RUN apt-get update && apt-get install -y&#x20;
tesseract-ocr&#x20;
tesseract-ocr-jav&#x20;
ca-certificates&#x20;
&& apt-get clean&#x20;
&& rm -rf /var/lib/apt/lists/\*

# Create app directory
WORKDIR /app

# for some reason, my 'out' dir is not get copied by docker
RUN mkdir -p /app/out

# Copy the Go binary from builder
COPY --from=builder /app/main .

# Copy static and templates if needed
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static

# Expose the port your app runs on
EXPOSE 8000

# Run the app
CMD \["./main"]
