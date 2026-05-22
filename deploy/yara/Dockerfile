FROM alpine:latest
RUN apk add --no-cache docker-cli
RUN apk add --no-cache yara
WORKDIR /data
ENTRYPOINT ["yara"]
CMD ["--help"]
