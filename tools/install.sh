#!/usr/bin/bash
set -xe
adduser --system fieldseeker-sync
cp bin/full_export /usr/local/bin/fieldseeker-sync-export
cp bin/webserver /usr/local/bin/fieldseeker-sync-webserver
cp systemd/fieldseeker-sync-export.service /etc/systemd/system/fieldseeker-sync-export.service
cp systemd/fieldseeker-sync-export.timer /etc/systemd/system/fieldseeker-sync-export.timer
cp systemd/fieldseeker-sync-webserver.service /etc/systemd/system/fieldseeker-sync-webserver.service
echo "Installation complete"
