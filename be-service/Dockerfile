FROM golang:alpine

# Set the working directory
WORKDIR /app

# # Copy go mod and sum files 
# COPY go.mod go.sum ./

# # Download dependencies  
# RUN go mod download 

# Copy the source from the current directory
COPY . /app

# Build the Go app
RUN go build -o main .

# Expose port 8001
EXPOSE 8001 

# Run the app
CMD ["/app/main"]