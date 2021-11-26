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
RUN go build


FROM nginx:stable-alpine
RUN apk --no-cache add busybox sqlite
COPY --from=builder /gwg/cmd/gwgather/gwgather /usr/local/bin/gwgather
COPY --from=builder /gwg/gwgather-config.json /etc/gwgather-config.json
COPY --from=builder /gwg/assets/dockerstart.sh /usr/local/bin/dockerstart.sh
CMD ["/usr/local/bin/dockerstart.sh"]
