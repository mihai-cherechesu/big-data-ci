FROM --platform=$BUILDPLATFORM golang:1.19-alpine as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ENV GO111MODULE=on

WORKDIR /app

# copy the files required for go mod download
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# copy all .go src files in the workdir
COPY internal ./internal
COPY *.go ./

# cross-compile the app for multi-arch support
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /controller

FROM --platform=$TARGETPLATFORM alpine
COPY --from=build /controller /controller

# just for documentation purposes, port is not published by this command
# 8081 is hardcoded in the source code at the moment
EXPOSE 8081

CMD ["/controller"]
