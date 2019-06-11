FROM golang:1.12-alpine
RUN apk --no-cache add ca-certificates make git
WORKDIR /app
COPY . .
RUN make && rm -rf vendor/

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /app/bin/docker-registry-pruner /bin/docker-registry-pruner
COPY ./entrypoint.sh /bin/entrypoint.sh
WORKDIR /app
COPY ./config ./config
ENTRYPOINT ["entrypoint.sh"]
CMD ["-h"]
