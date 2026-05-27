FROM alpine
LABEL authors="mephalrith"

WORKDIR /config

COPY ficsit_linux_amd64.apk ./
RUN apk add --allow-untrusted ficsit_linux_amd64.apk && \
    rm ficsit_linux_amd64.apk

ENTRYPOINT ["top", "-b"]