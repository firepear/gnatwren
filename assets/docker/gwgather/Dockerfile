# gwgather dockerfile
#
# To build: 'docker build --tag gwgather .'
#
# To launch: docker run --rm -p 9099:80 -v /PATH/TO/homefarm:/homefarm -it control /bin/bash
#

FROM golang:alpine
RUN apk --no-cache add gcc musl-dev
WORKDIR /gwg
COPY . /gwg/
WORKDIR cmd/gwgather
RUN go build


FROM golang:alpine
COPY --from=0 /gwg/cmd/gwgather/gwgather ./
CMD ["./gwgather", "-config", "gwgather-config.json"]
