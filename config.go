package main

import (
	"github.com/spf13/viper"
)

type ConfigArcgis struct {
	ServiceRoot string
	TenantID    string
	Token       string
}
type ConfigDatabase struct {
	URL string
}
type Config struct {
	Arcgis   ConfigArcgis
	Database ConfigDatabase
}

func ReadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("fieldseeker-sync-bridge") // name of config file (without extension)
	v.SetConfigType("toml")                    // REQUIRED if the config file does not have the extension in the name
	v.AddConfigPath("/etc/")                   // path to look for the config file in
	v.AddConfigPath("$HOME/.config")           // call multiple times to add many search paths
	v.AddConfigPath(".")                       // optionally look for config in the working directory
	err := v.ReadInConfig()                    // Find and read the config file
	if err != nil {                            // Handle errors reading the config file
		return nil, err
	}
	var c Config

	err = v.Unmarshal(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
