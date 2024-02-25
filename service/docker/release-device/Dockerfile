# syntax=docker/dockerfile:1

FROM golang:1.21.5
ARG MACHINE_NAME="hmi"
ARG REGISTRY_URL=seil-nexus-rnd.solaredge.com:27102

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY cmd/ ./cmd
COPY pkg/ ./pkg
COPY Makefile ./
COPY version ./

# Build # TODO replace this!!!!!!!!!!!!!!!!!!!!!!!!!!1
RUN go mod edit -replace gitlab.solaredge.com/utility/edge-hub/utility-edge-metadata.git=/tmp/source
# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose
#EXPOSE 8080

#RUN APP_VERSION=$(shell cat ./version) && GOOS=linux GOARCH=arm64 GOARM=7 go build -o=./build/target/arm64/network-manager -ldflags="-X 'main.Version=${APP_VERSION}' -s -w" ./cmd/network-manager
ENV GOPROXY="http://docker-esh.solaredge.com/artifactory/api/go/utility-go"
ENV GOPRIVATE=gitlab.solaredge.com
ENV GONOSUMDB="gitlab.solaredge.com/portialinuxdevelopers/*"
ENV GONOPROXY=github.com/*,golang.org/*

# CMD ["sh"]
# RUN go mod tidy
# RUN GOOS=linux GOARCH=arm64 GOARM=7 go build -o=./build/target/arm64/network-manager -ldflags="-X 'main.Version=1.1.1' -s -w" ./cmd/${MACHINE_NAME}

## Run
#RUN ls
#CMD ["sh"]
