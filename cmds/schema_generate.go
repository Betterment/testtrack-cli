package cmds

import (
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/spf13/cobra"
)

var schemaGenerateDoc = `
Reads the migrations in testtrack/migrate and writes the resulting schema state
to testtrack/schema.{json,yml}, overwriting the file if it already exists. Generate
makes no TestTrack API calls.

In addition to refreshing a schema file that may have been corrupted due to
a bad merge or bug that produced incorrect schema state, 'schema generate' will
also validate that migrations merged from multiple development branches don't
logically conflict, or else it will fail with errors.

Note that the generated schema is not guaranteed to match TestTrack server
state if multiple migrations affecting the same resource (e.g. split) were
applied out of timestamp order due to merge order.

You can realign your TestTrack server configuration with a generated schema
file by running the 'schema load' command, but be aware that decisions made in
the TestTrack admin may be overridden as a result.
`

func init() {
	schemaCmd.AddCommand(schemaGenerateCmd)
}

var schemaGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate schema.{json,yml} from migration files",
	Long:  schemaGenerateDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return schemaGenerate()
	},
}

func schemaGenerate() error {
	_, err := schema.Generate()
	return err
}
