package fssync

import (
	"os"
)

type ConfigArcgis struct {
	FieldSeekerService string
	ServiceRoot        string
	TenantID           string
	Token              string
}
type ConfigDatabase struct {
	URL string
}
type ConfigUserFiles struct {
	Directory string
}
type ConfigWebhook struct {
	Secret string
}

type Config struct {
	Arcgis   ConfigArcgis
	Database ConfigDatabase
	UserFiles ConfigUserFiles
	Webhook  ConfigWebhook
}

func ReadConfig() (*Config, error) {
	var c Config
	c.Arcgis.FieldSeekerService = os.Getenv("FIELDSEEKER_SYNC_ARCGIS_FIELDSEEKERSERVICE")
	c.Arcgis.ServiceRoot = os.Getenv("FIELDSEEKER_SYNC_ARCGIS_SERVICEROOT")
	c.Arcgis.TenantID = os.Getenv("FIELDSEEKER_SYNC_ARCGIS_TENANTID")
	c.Arcgis.Token = os.Getenv("FIELDSEEKER_SYNC_ARCGIS_TOKEN")
	c.Database.URL = os.Getenv("FIELDSEEKER_SYNC_DATABASE_URL")
	c.UserFiles.Directory = os.Getenv("FIELDSEEKER_SYNC_USERFILES_DIRECTORY")
	if len(c.UserFiles.Directory) == 0 {
		c.UserFiles.Directory = "/opt/fieldseeker-sync/data"
	}
	c.Webhook.Secret = os.Getenv("FIELDSEEKER_SYNC_WEBHOOK_SECRET")
	return &c, nil
}
