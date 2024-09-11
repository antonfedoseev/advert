#!/bin/bash

set -e

DIR=$(cd "$(dirname "$0")"; pwd -P)

docker-compose -f "$DIR/mysql.yml" -f "$DIR/redis.yml" -p "pet-advert-service" up -d "$@"