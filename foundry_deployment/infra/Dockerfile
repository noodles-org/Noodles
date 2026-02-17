FROM debian:bookworm

WORKDIR /

ARG FOUNDRY_URL

RUN apt-get update && apt-get install -y zip unzip curl
RUN apt-get install -y libssl-dev && \
    curl -sL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs
RUN mkdir -p /foundryvtt /foundrydata

COPY playbooks/scripts/foundry_start.sh /usr/local/bin/
RUN chmod u+rwx /usr/local/bin/foundry_start.sh

RUN cd /foundryvtt && \
    curl -Lo foundryvtt-node.zip "${FOUNDRY_URL}" && \
    unzip -o foundryvtt-node.zip && \
    rm foundryvtt-node.zip

EXPOSE 30000

ENTRYPOINT ["foundry_start.sh"]
