FROM alpine:3.12

COPY ttn-lw-migrate /bin/ttn-lw-migrate

ENTRYPOINT ["ttn-lw-migrate"]
