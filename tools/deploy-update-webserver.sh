#!/usr/bin/bash
set -xe
systemctl stop fieldseeker-sync-webserver.service
cp bin/webserver /usr/local/bin/fieldseeker-sync-webserver
systemctl start fieldseeker-sync-webserver.service
