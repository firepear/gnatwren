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


FROM nginx
COPY --from=builder /gwg/cmd/gwgather/gwgather /gwg/gather-config.json ./
CMD ["./gwgather", "-config", "gwgather-config.json"]
