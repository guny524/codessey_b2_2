FROM node:22-slim
RUN apt update \
    && apt install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && npm install -g @openai/codex
COPY image_proxy /usr/local/bin/image_proxy
ENV IMAGE_PROXY_ADDR=0.0.0.0:8080
EXPOSE 8080
ENTRYPOINT ["image_proxy"]
