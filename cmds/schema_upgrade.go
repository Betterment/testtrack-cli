package cmds

import (
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/spf13/cobra"
)

var schemaUpgradeDoc = `
Upgrades testtrack/schema.{json,yml} to the schema format this CLI writes,
in place, without replaying migrations.

It keeps the materialized state already in the file (splits, decisions,
retirements, feature completions, identifier types) and only rebuilds the
applied-migration version list from the files in testtrack/migrate. This is the
upgrade path to use when a newer CLI refuses to read an older-format schema.

Unlike 'schema generate', upgrade does not rebuild the schema from scratch, so
it works on apps whose migrations can't be replayed cleanly - e.g. when
testtrack/migrate predates some splits, or references splits that were created
out of band in the TestTrack admin.
`

func init() {
	schemaCmd.AddCommand(schemaUpgradeCmd)
}

var schemaUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade schema.{json,yml} to the current format in place",
	Long:  schemaUpgradeDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return schemaUpgrade()
	},
}

func schemaUpgrade() error {
	_, err := schema.Upgrade()
	return err
}
