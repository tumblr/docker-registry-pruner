# Hacking

## Building

### Local Tooling

To build with a local install of `go`:

```
$ make
▶ running gofmt…
▶ running golint…
▶ building executable…
Built bin/docker-registry-pruner: v0.1.0-35-ga4063b0-dirty 2019-06-10T16:43:02-0400
$ ./bin/docker-registry-pruner -h
Usage of ./bin/docker-registry-pruner:
docker-registry-pruner version:v0.1.0-35-ga4063b0-dirty (commit:a4063b02219053e8bfcb7ae6d457ce5c32c6e0cd branch:master) built on 2019-06-10T16:43:02-0400 with go1.12.3
  -config string
        Config yaml (default "config.yaml")
  -mode string
        Select operation mode (default "report")
```

### Docker

```
▶ building docker container…
docker build -t tumblr/docker-registry-pruner:v0.1.0-35-ga4063b0-dirty .
Sending build context to Docker daemon  11.55MB
Step 1/13 : FROM golang:1.12-alpine
 ---> c0e5aac9423b
Step 2/13 : RUN apk --no-cache add ca-certificates make git
 ---> Using cache
 ---> 74cbba72b7c4
Step 3/13 : WORKDIR /app
 ---> Using cache
 ---> 0be25eeae010
Step 4/13 : COPY . .
 ---> Using cache
 ---> 4f13727ab1e7
Step 5/13 : RUN make && rm -rf vendor/
 ---> Using cache
 ---> 0aabb9b14315
Step 6/13 : FROM alpine:latest
 ---> 3f53bb00af94
Step 7/13 : RUN apk --no-cache add ca-certificates
 ---> Using cache
 ---> b7d3d4b484a9
Step 8/13 : COPY --from=0 /app/bin/docker-registry-pruner /bin/docker-registry-pruner
 ---> Using cache
 ---> c6d2999d37a0
Step 9/13 : COPY ./entrypoint.sh /bin/entrypoint.sh
 ---> Using cache
 ---> 72c24b111702
Step 10/13 : WORKDIR /app
 ---> Using cache
 ---> ad0919ff365c
Step 11/13 : COPY ./config ./config
 ---> Using cache
 ---> 35806f6537ed
Step 12/13 : ENTRYPOINT ["entrypoint.sh"]
 ---> Using cache
 ---> b7ced5eef683
Step 13/13 : CMD ["-h"]
 ---> Using cache
 ---> 4e123cca3999
Successfully built 4e123cca3999
Successfully tagged tumblr/docker-registry-pruner:latest
$ docker run -it tumblr/docker-registry-pruner:latest -h
+ exec docker-registry-pruner -h
Usage of docker-registry-pruner:
docker-registry-pruner version:v0.1.0-35-ga4063b0-dirty (commit:a4063b02219053e8bfcb7ae6d457ce5c32c6e0cd branch:master) built on 2019-06-10T20:45:07+0000 with go1.12.1
  -config string
        Config yaml (default "config.yaml")
  -mode string
        Select operation mode (default "report")
```

## Opening a PR

To open a PR, please make sure you:

* Fill in the PR template
* Provide rationale about what improvement/bug you are addressing (link to Github issue too!)
* Provide testing methodology, expected behavior, and unit tests
* @byxorna for review

