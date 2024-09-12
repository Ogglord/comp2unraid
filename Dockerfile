# Use an official Go image as the base
FROM --platform=$BUILDPLATFORM golang:alpine
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG GIT_COMMIT
ARG GIT_BRANCH
ARG IMAGE_NAME
ENV GIT_COMMIT=$GIT_COMMIT


RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"
RUN echo "This is commit:$GIT_COMMIT for ${IMAGE_NAME}/$GIT_BRANCH"
# Set the working directory to /app
WORKDIR /app

# Copy the Go mod files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the application code
COPY . .
COPY entrypoint.sh .

# Build the application and inject the git commit
RUN go build -o comp2unraid

RUN chmod +x ./entrypoint.sh


ENTRYPOINT [ "/app/entrypoint.sh" ]

# Run the command to start the application
CMD ["-h"]