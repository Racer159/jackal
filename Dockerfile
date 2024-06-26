FROM cgr.dev/chainguard/static:latest
ARG TARGETARCH

# 65532 is the UID of the `nonroot` user in chainguard/static.  See: https://edu.chainguard.dev/chainguard/chainguard-images/reference/static/overview/#users
USER 65532:65532

COPY --chown=65532:65532 "build/jackal-linux-$TARGETARCH" /jackal

CMD ["/jackal", "internal", "agent", "-l=trace", "--no-log-file"]
