#!/bin/bash

set -e

DIR=$(cd "$(dirname "$0")"; pwd -P)

docker-compose -f "$DIR/advertd.yml" -p "pet-advert-service" up --build -d "$@"