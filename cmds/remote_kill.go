package cmds

import (
	"github.com/Betterment/testtrack-cli/migrationmanagers"
	"github.com/Betterment/testtrack-cli/remotekills"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/spf13/cobra"
)

var remoteKillDoc = `
Sets or updates a split remote-kill for a range of app versions, forcing all
users of affected apps to see the override_to variant of the specified split
between first_bad_version and an optional fixed_version.

Example:

testtrack remote_kill my_fancy_experiment catastrophic_bug_jan_2019 control 1.0 1.1

Reason should be a camel_case slug.

Submitting another remote_kill with the same reason will modify the existing
remote_kill.

Override-to is the variant affected app users should see.

Server-side apps will typically ignore this setting and show features
regardless of remote kill state because they can simply decide the split until
the bug can be fixed and then undecide it afterward.

You can reverse remote_kills with the delete_remote_kill command.
`

func init() {
	rootCmd.AddCommand(remoteKillCmd)
}

var remoteKillCmd = &cobra.Command{
	Use:   "remote_kill split_name reason override_to first_bad_version [fixed_version]",
	Short: "Set or update a split remote-kill for a range of app versions",
	Long:  remoteKillDoc,
	Args:  cobra.RangeArgs(4, 5),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 5 {
			return remoteKill(&args[0], &args[1], &args[2], &args[3], &args[4])
		}
		return remoteKill(&args[0], &args[1], &args[2], &args[3], nil)
	},
}

func remoteKill(split, reason, overrideTo, firstBadVersion, fixedVersion *string) error {
	// This validation is the difference between remote_kill and unset_remote_kill which is why it's inline
	err := validations.Presence("first_bad_version", firstBadVersion)
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
