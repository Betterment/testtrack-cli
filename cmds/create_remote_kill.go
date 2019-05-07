package cmds

import (
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/remotekills"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var createRemoteKillDoc = `
Sets or updates a split remote-kill for a range of app versions, forcing all
users of affected apps to see the override_to variant of the specified split
between first_bad_version and an optional fixed_version.

Example:

testtrack create remote_kill my_fancy_experiment catastrophic_bug_jan_2019 --override_to control --first_bad_version 1.0 --fixed_version 1.1

Reason should be a camel_case slug.

Submitting another remote_kill with the same reason will modify the existing
remote_kill.

Override-to is the variant affected app users should see.

Server-side apps will typically ignore this setting and show features
regardless of remote kill state because they can simply decide the split until
the bug can be fixed and then undecide it afterward.

You can reverse remote_kills with the destroy remote_kill command.
`

var createRemoteKillOverrideTo, createRemoteKillFirstBadVersion, createRemoteKillFixedVersion string

func init() {
	createRemoteKillCmd.Flags().StringVar(&createRemoteKillOverrideTo, "override_to", "", "Override-to variant (required)")
	createRemoteKillCmd.MarkFlagRequired("override_to")
	createRemoteKillCmd.Flags().StringVar(&createRemoteKillFirstBadVersion, "first_bad_version", "", "First bad app version (required)")
	createRemoteKillCmd.MarkFlagRequired("first_bad_version")
	createRemoteKillCmd.Flags().StringVar(&createRemoteKillFixedVersion, "fixed_version", "", "Fixed app version")
	createRemoteKillCmd.Flags().BoolVar(&noPrefix, "no-prefix", false, "Don't prefix split with app_name to refer to legacy splits")
	createRemoteKillCmd.Flags().BoolVar(&force, "force", false, "Force creation if split isn't found in schema, e.g. if split is retired")
	createCmd.AddCommand(createRemoteKillCmd)
}

var createRemoteKillCmd = &cobra.Command{
	Use:   "remote_kill split_name reason",
	Short: "Set or update a split remote-kill for a range of app versions",
	Long:  createRemoteKillDoc,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createRemoteKill(&args[0], &args[1], &createRemoteKillOverrideTo, &createRemoteKillFirstBadVersion, &createRemoteKillFixedVersion)
	},
}

func createRemoteKill(split, reason, overrideTo, firstBadVersion, fixedVersion *string) error {
	currentAppName, err := getAppName()
	if err != nil {
		return err
	}
	schema, err := schema.Read()
	if err != nil {
		return err
	}
	err = validations.AutoPrefixAndValidateSplit("split_name", split, currentAppName, schema, noPrefix, force)
	if err != nil {
		return err
	}

	// These validations are the difference between create and destroy remote_kill which is why they're inline
	err = validations.Presence("override_to", overrideTo)
	if err != nil {
		return err
	}
	err = validations.Presence("first_bad_version", firstBadVersion)
	if err != nil {
		return err
	}

	remoteKill, err := remotekills.New(split, reason, overrideTo, firstBadVersion, fixedVersion)
	if err != nil {
		return err
	}

	mgr, err := migrationmanagers.New(remoteKill)
	if err != nil {
		return err
	}

	err = mgr.Save()
	if err != nil {
		return err
	}

	return nil
}
