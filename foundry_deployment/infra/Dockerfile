FROM node:26-slim

WORKDIR /

ARG FOUNDRY_URL

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl zip unzip ca-certificates && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /foundryvtt /foundrydata

COPY playbooks/scripts/foundry_start.sh /usr/local/bin/
RUN chmod u+rwx /usr/local/bin/foundry_start.sh

RUN cd /foundryvtt && \
    curl -Lo foundryvtt-node.zip "${FOUNDRY_URL}" && \
    unzip -o foundryvtt-node.zip && \
    rm foundryvtt-node.zip

EXPOSE 30000

ENTRYPOINT ["foundry_start.sh"]
