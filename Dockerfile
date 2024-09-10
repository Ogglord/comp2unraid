# Use an official Go image as the base
FROM --platform=$BUILDPLATFORM golang:alpine
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"

# Set the working directory to /app
WORKDIR /app

# Copy the Go mod files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the application code
COPY . .
COPY entrypoint.sh .

# Build the application
RUN go build -o comp2unraid main.go

RUN chmod +x ./entrypoint.sh


ENTRYPOINT [ "/app/entrypoint.sh" ]

# Run the command to start the application
CMD ["-h"]