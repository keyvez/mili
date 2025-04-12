FROM golang:1.24.2-alpine3.21 as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
# Build the binary.
RUN go build -mod=readonly -v -o main

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:3.21.0
# RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
#     ca-certificates && \
#     rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/main /app/main

# Run the web service on container startup.
CMD ["/app/main"]