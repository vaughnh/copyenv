package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

type CopyEnv struct{}

func fatalIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, "error: ", err)
		os.Exit(1)
	}
}

func (c *CopyEnv) ExtractAppName(args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("missing application name")
	}

	return args[1], nil
}

func (c *CopyEnv) RetrieveAppNameEnv(cliConnection plugin.CliConnection, app_name string) ([]string, error) {
	output, err := cliConnection.CliCommandWithoutTerminalOutput("env", app_name)

	if err != nil {
		for _, val := range output {
			fmt.Println(val)
		}
	}

	return output, err
}

func (c *CopyEnv) ExtractCredentialsJSON(credKey string, output []string) ([]byte, error) {
	err := errors.New("missing service credentials for application")
	var b []byte

	for _, val := range output {
		if strings.Contains(val, credKey) {
			var f interface{}
			err = json.Unmarshal([]byte(val), &f)
			if err != nil {
				return nil, err
			}

			m := f.(map[string]interface{})
			b, err = json.Marshal(m[credKey])
			if err != nil {
				return nil, err
			}

		}
	}

	return b, err
}

func (c *CopyEnv) ExportCredsAsShellVar(credKey string, creds string) {
	vcap_services := fmt.Sprintf("export %s='%s';", credKey, creds)
	fmt.Println(vcap_services)
}

func (c *CopyEnv) Run(cliConnection plugin.CliConnection, args []string) {
	if len(args) > 0 && args[0] == "CLI-MESSAGE-UNINSTALL" {
		return
	}
	app_name, err := c.ExtractAppName(args)
	fatalIf(err)

	app_env, err := c.RetrieveAppNameEnv(cliConnection, app_name)
	fatalIf(err)

	services_creds, err := c.ExtractCredentialsJSON("VCAP_SERVICES", app_env)
	fatalIf(err)

	c.ExportCredsAsShellVar("VCAP_SERVICES", string(services_creds[:]))
}

func (c *CopyEnv) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "copyenv",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			plugin.Command{
				Name:     "copyenv",
				HelpText: "Export application VCAP_SERVICES to local environment variable.",
				UsageDetails: plugin.Usage{
					Usage: "copyenv APP_NAME - Retrieve and export remote application VCAP_SERVICES to local developer environment.",
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(CopyEnv))
}
