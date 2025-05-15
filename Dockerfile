FROM golang:1.24

# set work dir
WORKDIR /app

# copy files from host to container
COPY . . 

# build the app inside container
RUN go build -o main . 

# Install Tesseract and Javanese language pack
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-jav \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# create 'out' directory
RUN mkdir -p /app/out 

# expose the server port
EXPOSE 8000 

# run the binary
CMD ["./main"] 
