#!/bin/sh
# make our cron dir if it doesn't exist
mkdir -p /etc/periodic/2min
# dump our script to be run and chmod it
cat > /etc/periodic/2min/gwdump <<EOF
#!/bin/sh
/usr/local/bin/gwdump -config /etc/gwgather-config.json
EOF
chmod +x /etc/periodic/2min/gwdump
# add the dir to the system crontab
echo "*/2 * * * *       run-parts /etc/periodic/2min" >> /etc/crontabs/root
# turn on the cron service
crond &
# background nginx
nginx -g 'daemon off;' &
# start gwgather
/usr/local/bin/gwgather -config /etc/gwgather-config.json
