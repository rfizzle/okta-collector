package main

import (
	"errors"
	"fmt"
	"github.com/rfizzle/collector-helpers/outputs"
	"github.com/rfizzle/collector-helpers/state"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func setupCliFlags() error {
	viper.SetEnvPrefix("OC")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	flag.Int("schedule", 30, "time in seconds to collect")
	flag.String("okta-domain", "", "okta domain for organization")
	flag.String("okta-api-key", "", "okta api key for authentication")
	flag.BoolP("verbose", "v", false, "verbose logging")
	flag.BoolP("config", "c", false, "enable config file")
	flag.String("config-path", "", "config file path")
	state.InitCLIParams()
	outputs.InitCLIParams()
	flag.Parse()
	err := viper.BindPFlags(flag.CommandLine)

	if err != nil {
		log.Fatalf("Failed parsing flags: %v", err.Error())
	}

	// Check config
	if err := checkConfigParams(); err != nil {
		return err
	}

	// Check parameters
	if err := checkRequiredParams(); err != nil {
		return err
	}

	return nil
}

func checkConfigParams() error {
	if viper.GetBool("config") {
		if !fileExists(viper.GetString("config-path")) {
			return errors.New("missing config file path param (--config-path)")
		}

		dir, file := filepath.Split(viper.GetString("config-path"))
		ext := strings.ToLower(filepath.Ext(viper.GetString("config-path")))

		supportedTypes := []string{"json", "toml", "yaml", "yml", "properties", "props", "prop", "env", "dotenv"}
		if !contains(supportedTypes, ext) {
			e := fmt.Sprintf("invalid config file type (supported: %s )", strings.Join(supportedTypes[:], ", "))
			return errors.New(e)
		}

		fileName := strings.TrimSuffix(file, ext)

		viper.SetConfigName(fileName)
		viper.SetConfigType(ext)
		viper.AddConfigPath(dir)

		err := viper.ReadInConfig() // Find and read the config file
		if err != nil { // Handle errors reading the config file
			return fmt.Errorf("Fatal error config file: %s \n", err)
		}
	}

	return nil
}

func checkRequiredParams() error {
	if viper.GetString("okta-domain") == "" {
		return errors.New("missing okta domain param (--okta-domain)")
	}

	if viper.GetString("okta-api-key") == "" {
		return errors.New("missing okta api key param (--okta-api-key)")
	}

	if err := state.ValidateCLIParams(); err != nil {
		return err
	}

	if err := outputs.ValidateCLIParams(); err != nil {
		return err
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}