package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/spf13/pflag"

	"code.storageos.net/storageos/c2-cli/apiclient"
	"code.storageos.net/storageos/c2-cli/apiclient/openapi"
	"code.storageos.net/storageos/c2-cli/cmd"
	"code.storageos.net/storageos/c2-cli/config"
	"code.storageos.net/storageos/c2-cli/config/environment"
	"code.storageos.net/storageos/c2-cli/config/flags"
)

var (
	// Version is the semantic version string which has been assigned to the
	// cli application.
	Version string
	// UserAgentPrefix is used by the CLI application to identify itself to
	// StorageOS.
	UserAgentPrefix string = "storageos-cli"
)

func main() {
	// Determine the cli app version from the embedded semver.
	// If it errors 0.0.0 is returned
	version, _ := semver.Make(Version)

	userAgent := strings.Join([]string{UserAgentPrefix, version.String()}, "/")

	// Initialise the configuration provider stack:
	//
	// → flags first. note we init the flagset here because it needs to be
	// given to the InitCommand call (setting up global flags)
	globalFlags := pflag.NewFlagSet("storageos", pflag.ContinueOnError)
	configProvider := flags.NewProvider(
		globalFlags,
		// → environment next
		environment.NewProvider(
			// → TODO(CP-3918) config file next
			//
			// → default values as final fallback
			config.NewDefaulter(),
		),
	)

	app := cmd.InitCommand(
		// Provide a closure for initialising the API client to the command,
		// so that it can be configured after flags have been parsed.
		func() (*apiclient.Client, error) {
			transport, err := openapi.NewOpenAPI(configProvider, userAgent)
			if err != nil {
				return nil, err
			}
			return apiclient.New(
				transport,
				configProvider,
			), nil
		},
		configProvider,
		globalFlags,
		version,
	)

	if err := app.Execute(); err != nil {
		// Attempt to map err to a command error.
		err = cmd.MapCommandError(err)
		// Get the appropriate exit code for the error.
		code := cmd.ExitCodeForError(err)

		fmt.Fprintf(app.OutOrStderr(), "Error: %v\n", err)

		os.Exit(code)
	}
}
