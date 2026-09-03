FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/bot .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/bot /bin/bot
ENTRYPOINT ["/bin/bot"]
