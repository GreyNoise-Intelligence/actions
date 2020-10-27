FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir -p /workspace

ENV ATHENA_VERSION=0.1.0

RUN wget -O /workspace/athena https://github.com/GreyNoise-Intelligence/actions/releases/download/v${ATHENA_VERSION}/athena-linux-amd64 && \
    chmod +x /workspace/athena

ENTRYPOINT ["/workspace/athena"]
