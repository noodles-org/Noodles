FROM debian:bookworm

WORKDIR /

RUN apt-get update && apt-get install -y zip unzip curl
RUN apt-get install -y libssl-dev && \
    curl -sL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs
RUN mkdir -p /foundryvtt /foundrydata

COPY playbooks/scripts/foundry_start.sh /usr/local/bin/
COPY FoundryVTT-Node-13.346.zip /foundryvtt

RUN cd /foundryvtt && \
    unzip -o FoundryVTT-Node-13.346.zip && \
    rm FoundryVTT-Node-13.346.zip && \
    chmod u+rwx /usr/local/bin/foundry_start.sh

EXPOSE 30000

ENTRYPOINT ["foundry_start.sh"]
