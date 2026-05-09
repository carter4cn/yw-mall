#!/bin/sh
# Replace localhost addresses with Docker Compose service hostnames before starting.
sed -i 's|127\.0\.0\.1:2379|etcd1:2379|g'       /app/etc/*.yaml
sed -i 's|127\.0\.0\.1:6033|proxysql:6033|g'     /app/etc/*.yaml
sed -i 's|127\.0\.0\.1:6379|redis-master:6379|g' /app/etc/*.yaml
sed -i 's|127\.0\.0\.1:9000|minio:9000|g'        /app/etc/*.yaml
exec /app/server "$@"
