#!/bin/sh
nginx -g 'daemon off;' &
/usr/local/bin/gwgather -config /etc/gwgather-config.json
