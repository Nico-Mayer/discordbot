FROM alpine:3.23

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 bot
COPY $TARGETOS/$TARGETARCH/bot /bin/bot
USER bot
ENTRYPOINT ["/bin/bot"]
