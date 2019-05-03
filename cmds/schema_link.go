package cmds

import (
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/spf13/cobra"
)

var schemaLinkDoc = `
Linking your schema allows 'testtrack server' to serve splits from your app.
This is an important part of configuring your app for local testtrack
development.

It's a great idea to call 'testtrack schema link' as part of a setup script
that developers call after cloning your app repo that installs dependencies,
provisions databases, etc.
`

func init() {
	schemaCmd.AddCommand(schemaLinkCmd)
}

var schemaLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Install schema into local testtrack server",
	Long:  schemaLinkDoc,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return schemaLink()
	},
}

func schemaLink() error {
	return schema.Link()
}
