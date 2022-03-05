#!/bin/bash

set -e

cd /opt/htm-docker
BACKUP_NAME=backup_$(date +"%Y_%m_%d_%H_%I_%S")
tar czf $BACKUP_NAME.tgz addon
mv $BACKUP_NAME.tgz /opt/htm-docker/backup/
find /opt/htm-docker/backup -type f -mtime +7 -delete
