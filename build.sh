#!/bin/bash
dockercmd=$(which docker 2>&1 || true)
if [[ "${dockercmd}" =~ ^which ]]; then
    dockercmd=$(which podman 2>&1 || true)
fi
if [[ "${dockercmd}" =~ ^which ]]; then
    echo "error: neither docker or podman found in PATH; bailing"
    exit 2
fi

${dockercmd} container stop gwgather && ${dockercmd} container rm gwgather && ${dockercmd} image rm gwgather
${dockercmd} image prune -f
${dockercmd} build --tag gwgather .
${dockercmd} volume create gwg
${dockercmd} run --name gwgather -d --restart always -p 9098:80 -p 11099:11099 -v gwg:/usr/share/nginx/html gwgather
${dockercmd} cp assets/web/index.html gwgather:/usr/share/nginx/html/
${dockercmd} cp assets/web/main.js gwgather:/usr/share/nginx/html/

