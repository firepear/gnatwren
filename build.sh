#!/bin/sh
docker container stop gwgather && docker container rm gwgather && docker image rm gwgather
docker image prune -f
docker build --tag gwgather .
docker volume create gwg
docker run --name gwgather -d --restart always -p 9098:80 -p 11099:11099 -v gwg:/usr/share/nginx/html gwgather
docker cp assets/web/index.html assets/web/main.js gwgather:/usr/share/nginx/html/

