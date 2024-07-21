#!/bin/bash

echo "db $(docker inspect --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' db)" >> /etc/hosts

exec "$@"
