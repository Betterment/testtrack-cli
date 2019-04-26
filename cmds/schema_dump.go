package cmds

import (
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/spf13/cobra"
)

var schemaDumpDoc = `
Reads the migrations in testtrack/migrate and dumps the resulting schema state
to testtrack/schema.yml

In addition to refreshing a schema.yml file that may have been corrupted due to
a bad merge or bug in testtrack that produced incorrect schema state, dumping
will also validate that migrations merged from multiple development branches
don't logically conflict, or else the dump will fail.

The dumped schema is not guaranteed to match production state if multiple
migrations affecting the same resource (e.g. split) were applied out of
timestamp order due to merge order.

You can put an environment back into alignment with a dumped schema file by
running the 'schema load' command.
`

func init() {
	schemaCmd.AddCommand(schemaDumpCmd)
}

var schemaDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump migrations to schema.yml",
	Long:  schemaDumpDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return schemaDump()
	},
}

func schemaDump() error {
	_, err := schema.Generate()
	return err
}
