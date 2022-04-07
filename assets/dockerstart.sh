#!/bin/sh
mkdir -p /etc/periodic/2min
cat > /etc/periodic/2min/gwdump <<EOF
#!/bin/sh
/usr/local/bin/gwdump -config /etc/gwgather-config.json
EOF
chmod +x /etc/periodic/2min/gwdump
echo "*/2 * * * *       run-parts /etc/periodic/2min" >> /etc/crontabs/root
nginx -g 'daemon off;' &
/usr/local/bin/gwgather -config /etc/gwgather-config.json
