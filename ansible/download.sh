#!/bin/bash

set -e

SCRIPT=`realpath $0`
ANSIBLE_PATH=`dirname $SCRIPT`

curl -s https://api.github.com/repos/cloudflare/cfssl/releases/latest | jq -r '.assets[].browser_download_url' | while read -r line; do
  if [[ $line == *"linux_amd64" ]]; then
    if [[ $line == *"cfssl_"* ]]; then
      wget -c -O $ANSIBLE_PATH/cert/cfssl "$line"
      chmod +x $ANSIBLE_PATH/cert/cfssl
    fi
    if [[ $line == *"cfssljson_"* ]]; then
      wget -c -O $ANSIBLE_PATH/cert/cfssljson "$line"
      chmod +x $ANSIBLE_PATH/cert/cfssljson
    fi
  fi
done

curl -s https://api.github.com/repos/aasaam/htm-node/releases/latest | jq -r '.assets[].browser_download_url' | while read -r line; do
  if [[ $line == *"linux.amd64.tgz" ]]; then
    wget -c -O $ANSIBLE_PATH/files/downloads/htm-node.tgz "$line"
  fi
done

docker pull ghcr.io/aasaam/web-server:latest
docker pull ghcr.io/aasaam/nginx-protection:latest
docker pull ghcr.io/aasaam/nginx-error-log-parser:latest
docker pull ghcr.io/aasaam/rest-captcha:latest
docker pull ghcr.io/aasaam/maxmind-lite-docker:latest

docker save ghcr.io/aasaam/web-server:latest -o $ANSIBLE_PATH//files/downloads/aasaam-web_server.tgz
docker save ghcr.io/aasaam/nginx-protection:latest -o $ANSIBLE_PATH//files/downloads/aasaam-nginx_protection.tgz
docker save ghcr.io/aasaam/nginx-error-log-parser:latest -o $ANSIBLE_PATH//files/downloads/aasaam-nginx_error_log_parser.tgz
docker save ghcr.io/aasaam/rest-captcha:latest -o $ANSIBLE_PATH//files/downloads/aasaam-rest_captcha.tgz
docker save ghcr.io/aasaam/maxmind-lite-docker:latest -o $ANSIBLE_PATH//files/downloads/aasaam-maxmind_lite_docker.tgz

docker rm -f maxmind-lite-docker-test
docker run --name maxmind-lite-docker-test -d ghcr.io/aasaam/maxmind-lite-docker tail -f /dev/null

docker cp maxmind-lite-docker-test:/GeoLite2-City.mmdb $ANSIBLE_PATH//files/downloads/GeoLite2-City.mmdb
docker cp maxmind-lite-docker-test:/GeoLite2-ASN.mmdb $ANSIBLE_PATH//files/downloads/GeoLite2-ASN.mmdb

docker rm -f maxmind-lite-docker-test

if [ ! -f $ANSIBLE_PATH/files/downloads/dhparam.pem ]; then
  openssl dhparam -out $ANSIBLE_PATH//files/downloads/dhparam.pem 2048
fi

if [ ! -f $ANSIBLE_PATH/cert/root.pem ]; then
  cd $ANSIBLE_PATH/cert
  ./cfssl gencert -initca csr-root.json | cfssljson -bare root
  ./cfssl gencert -initca csr-intermediate.json | cfssljson -bare intermediate
  ./cfssl sign -ca root.pem -ca-key root-key.pem -config ca-config.json -profile intermediate intermediate.csr | cfssljson -bare intermediate
fi
