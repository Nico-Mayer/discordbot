FROM gcr.io/distroless/static-debian13:nonroot

ARG TARGETOS
ARG TARGETARCH

COPY $TARGETOS/$TARGETARCH/bot /bin/bot
ENTRYPOINT ["/bin/bot"]
