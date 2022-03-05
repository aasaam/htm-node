#!/bin/bash

set -e

docker pull ghcr.io/aasaam/web-server:latest
docker pull ghcr.io/aasaam/nginx-protection:latest
docker pull ghcr.io/aasaam/nginx-error-log-parser:latest
docker pull ghcr.io/aasaam/rest-captcha:latest
docker pull ghcr.io/aasaam/maxmind-lite-docker:latest

docker save ghcr.io/aasaam/web-server:latest -o ./files/downloads/aasaam-web_server.tgz
docker save ghcr.io/aasaam/nginx-protection:latest -o ./files/downloads/aasaam-nginx_protection.tgz
docker save ghcr.io/aasaam/nginx-error-log-parser:latest -o ./files/downloads/aasaam-nginx_error_log_parser.tgz
docker save ghcr.io/aasaam/rest-captcha:latest -o ./files/downloads/aasaam-rest_captcha.tgz
docker save ghcr.io/aasaam/maxmind-lite-docker:latest -o ./files/downloads/aasaam-maxmind_lite_docker.tgz

docker rm -f maxmind-lite-docker-test
docker run --name maxmind-lite-docker-test -d ghcr.io/aasaam/maxmind-lite-docker tail -f /dev/null

docker cp maxmind-lite-docker-test:/GeoLite2-City.mmdb ./files/downloads/GeoLite2-City.mmdb
docker cp maxmind-lite-docker-test:/GeoLite2-ASN.mmdb ./files/downloads/GeoLite2-ASN.mmdb

docker rm -f maxmind-lite-docker-test

if [ ! -f ./files/downloads/dhparam.pem ]; then
  openssl dhparam -out ./files/downloads/dhparam.pem 2048
fi

