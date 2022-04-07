# gwgather dockerfile
#
# To build: 'docker build --tag gwgather .'
#
# To launch: docker run -d --restart always -p CONFIG_PORT:11099 -v /PATH/TO/DBDIR:/db
#

FROM golang:alpine as builder
RUN apk --no-cache add gcc musl-dev
WORKDIR /gwg
COPY . /gwg/
WORKDIR cmd/gwgather
ENV CGO_ENABLED=1 CGO_CFLAGS="-DSQLITE_ENABLE_JSON1"
RUN go build
WORKDIR /gwg/cmd/gwdump
RUN go build


FROM nginx:stable-alpine
RUN apk --no-cache add busybox sqlite
COPY --from=builder /gwg/cmd/gwgather/gwgather /gwg/cmd/gwdump/gwdump /gwg/assets/dockerstart.sh /usr/local/bin/
COPY --from=builder /gwg/assets/gwgather-config.json /etc/gwgather-config.json
CMD ["/usr/local/bin/dockerstart.sh"]
