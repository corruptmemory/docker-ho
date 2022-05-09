FROM alpine:latest

RUN mkdir -p /opt/program

COPY docker_out/docker-ho /opt/program

EXPOSE 8080

ENTRYPOINT ["/opt/program/docker-ho"]
