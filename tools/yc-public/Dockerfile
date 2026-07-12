FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends bash ca-certificates curl jq less \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://storage.yandexcloud.net/yandexcloud-yc/install.sh \
    | bash -s -- -i /opt/yandex-cloud -n

ENV PATH="/opt/yandex-cloud/bin:${PATH}"

WORKDIR /workspace
ENTRYPOINT ["yc"]
