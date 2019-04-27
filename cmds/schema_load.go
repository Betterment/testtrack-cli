package cmds

import (
	"github.com/Betterment/testtrack-cli/schemaloaders"
	"github.com/spf13/cobra"
)

var schemaLoadDoc = `
Loads the testtrack/schema.yml state into TestTrack server. This operation is
idempotent with a valid, consistent schema file, though might fail if your
schema file became invalid due to a bad merge or a bug.

If a schema fails to load, you can diagnose/fix by calling 'schema generate' to
regenerate the schema from the migrations on the filesystem. Migrations could
become inconsistent as well due to merges because of duplicative or conflicting
migrations impacting the same resource (e.g. split), but if so, generate will
fail and provide diagnostic error messages you can use to identify conflicts
and potentially delete the problematic migrations.

Be aware that decisions made in the TestTrack admin may be overridden by
running 'schema load'.
`

func init() {
	schemaCmd.AddCommand(schemaLoadCmd)
}

var schemaLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load schema.yml state into TestTrack server",
	Long:  schemaLoadDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return schemaLoad()
	},
}

func schemaLoad() error {
	loader, err := schemaloaders.New()
	if err != nil {
		return err
	}
	return loader.Load()
}
