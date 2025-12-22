package relay

import (
	"errors"
	"fmt"

	"github.com/abcxyz/pkg/cli"
)

// Config defines the set over environment variables required
// for running this application.
type Config struct {
	Port           string
	ProjectID      string
	RelayTopicID   string
	RelayProjectID string
}

// Validate validates the service config after load.
func (cfg *Config) Validate() error {
	var merr error

	if cfg.ProjectID == "" {
		merr = errors.Join(merr, fmt.Errorf("PROJECT_ID is required"))
	}

	return merr
}

// ToFlags binds the config to the give [cli.FlagSet] and returns it.
func (cfg *Config) ToFlags(set *cli.FlagSet) *cli.FlagSet {
	f := set.NewSection("RELAY OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "port",
		Target:  &cfg.Port,
		EnvVar:  "PORT",
		Default: "8080",
		Usage:   `The port the relay server listens to.`,
	})

	f.StringVar(&cli.StringVar{
		Name:   "project-id",
		Target: &cfg.ProjectID,
		EnvVar: "PROJECT_ID",
		Usage:  `Google Cloud project ID where this service runs.`,
	})
	f.StringVar(&cli.StringVar{
		Name:   "relay-topic-id",
		Target: &cfg.RelayTopicID,
		EnvVar: "RELAY_TOPIC_ID",
		Usage:  `Google PubSub topic ID.`,
	})
	f.StringVar(&cli.StringVar{
		Name:   "relay-project-id",
		Target: &cfg.RelayProjectID,
		EnvVar: "RELAY_PROJECT_ID",
		Usage:  `Google Cloud project ID where the relay topic lives.`,
	})

	return set
}
