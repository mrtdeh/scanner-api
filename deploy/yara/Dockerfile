FROM alpine:latest
RUN apk add --no-cache yara

WORKDIR /data
COPY ./deploy/yara/entrypoint.sh /data/entrypoint.sh
RUN chmod +x /data/entrypoint.sh


ENTRYPOINT ["/data/entrypoint.sh"]

