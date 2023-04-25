package validations_test

import (
	"os"
	"testing"

	"github.com/Betterment/testtrack-cli/serializers"
	"github.com/Betterment/testtrack-cli/validations"
	"github.com/stretchr/testify/require"
)

func TestAutoPrefixAndValidateSplit(t *testing.T) {
	t.Run("it blows up when noPrefix: true and input is prefixed", func(t *testing.T) {
		paramName := "split_name"
		value := StrPtr("my_app.foo_enabled")
		currentAppName := "my_app"
		schema := &serializers.Schema{}
		noPrefix := true
		force := false
		err := validations.AutoPrefixAndValidateSplit(paramName, value, currentAppName, schema, noPrefix, force)
		require.Error(t, err)
		require.Contains(t, err.Error(), "--no-prefix incompatible with prefix 'my_app'")
	})

	t.Run("its result is unprefixed when noPrefix: true, force: true, and input is unprefixed", func(t *testing.T) {
		paramName := "split_name"
		value := StrPtr("foo_enabled")
		currentAppName := "my_app"
		schema := &serializers.Schema{}
		noPrefix := true
		force := true
		err := validations.AutoPrefixAndValidateSplit(paramName, value, currentAppName, schema, noPrefix, force)
		require.NoError(t, err)
		require.Equal(t, "foo_enabled", *value)
	})

	t.Run("its result is prefixed when noPrefix: false, force: true, and input is prefixed", func(t *testing.T) {
		paramName := "split_name"
		value := StrPtr("my_app.foo_enabled")
		currentAppName := "my_app"
		schema := &serializers.Schema{}
		noPrefix := false
		force := true
		err := validations.AutoPrefixAndValidateSplit(paramName, value, currentAppName, schema, noPrefix, force)
		require.NoError(t, err)
		require.Equal(t, "my_app.foo_enabled", *value)
	})

	t.Run("its result is prefixed when noPrefix: false, force: true, and input is unprefixed", func(t *testing.T) {
		paramName := "split_name"
		value := StrPtr("foo_enabled")
		currentAppName := "my_app"
		schema := &serializers.Schema{}
		noPrefix := false
		force := true
		err := validations.AutoPrefixAndValidateSplit(paramName, value, currentAppName, schema, noPrefix, force)
		require.NoError(t, err)
		require.Equal(t, "my_app.foo_enabled", *value)
	})

	t.Run("its result is prefixed when noPrefix: false, force: false, and input is unprefixed and split exists prefixed in schema", func(t *testing.T) {
		paramName := "split_name"
		value := StrPtr("foo_enabled")
		currentAppName := "my_app"
		schema := &serializers.Schema{
			Splits: []serializers.SchemaSplit{
				serializers.SchemaSplit{
					Name: "my_app.foo_enabled",
				},
			},
		}
		noPrefix := false
		force := false
		err := validations.AutoPrefixAndValidateSplit(paramName, value, currentAppName, schema, noPrefix, force)
		require.NoError(t, err)
		require.Equal(t, "my_app.foo_enabled", *value)
	})

	t.Run("its result is prefixed when noPrefix: false, force: false, and input is prefixed and split exists prefixed in schema", func(t *testing.T) {
		paramName := "split_name"
		value := StrPtr("my_app.foo_enabled")
		currentAppName := "my_app"
		schema := &serializers.Schema{
			Splits: []serializers.SchemaSplit{
				serializers.SchemaSplit{
					Name: "foo_enabled",
				},
				serializers.SchemaSplit{
					Name: "my_app.foo_enabled",
				},
			},
		}
		noPrefix := false
		force := false
		err := validations.AutoPrefixAndValidateSplit(paramName, value, currentAppName, schema, noPrefix, force)
		require.NoError(t, err)
		require.Equal(t, "my_app.foo_enabled", *value)
	})

	t.Run("it blows up when noPrefix: true, force: false, and input is unprefixed and split doesn't exist in schema", func(t *testing.T) {
		paramName := "split_name"
		value := StrPtr("foo_enabled")
		currentAppName := "my_app"
		schema := &serializers.Schema{}
		noPrefix := true
		force := false
		err := validations.AutoPrefixAndValidateSplit(paramName, value, currentAppName, schema, noPrefix, force)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found in schema")
	})

	t.Run("its result is unprefixed when noPrefix: true, force: false, and input is unprefixed and split exists unprefixed in schema", func(t *testing.T) {
		paramName := "split_name"
		value := StrPtr("foo_enabled")
		currentAppName := "my_app"
		schema := &serializers.Schema{
			Splits: []serializers.SchemaSplit{
				serializers.SchemaSplit{
					Name: "my_app.foo_enabled",
				},
				serializers.SchemaSplit{
					Name: "foo_enabled",
				},
			},
		}
		noPrefix := true
		force := false
		err := validations.AutoPrefixAndValidateSplit(paramName, value, currentAppName, schema, noPrefix, force)
		require.NoError(t, err)
		require.Equal(t, "foo_enabled", *value)
	})
}

func TestValidateOwnerName(t *testing.T) {
	t.Run("it succeeds with no owner if ownershipFilename is undefined", func(t *testing.T) {
		err := validations.ValidateOwnerName("", "")
		require.NoError(t, err)
	})

	t.Run("it fails with an owner if ownershipFilename is undefined", func(t *testing.T) {
		err := validations.ValidateOwnerName("super_owner", "")
		require.Contains(t, err.Error(), "owner must be empty when TESTTRACK_OWNERSHIP_FILE is not defined")
	})

	t.Run("it succeeds without an owner if no ownership file is present", func(t *testing.T) {
		var ownershipFilename = ".owners.yml"

		owner := ""
		err := validations.ValidateOwnerName(owner, ownershipFilename)
		require.NoError(t, err)
	})

	t.Run("it errors out if an ownership file is present and the owner is blank", func(t *testing.T) {
		var ownershipFilename = ".owners.yml"

		ownerContent := []byte("super_squad:\n  delayed_job_alert_slack_channel: '#super_squad'\n")
		os.WriteFile(ownershipFilename, ownerContent, 0644)

		owner := ""
		err := validations.ValidateOwnerName(owner, ownershipFilename)

		os.Remove(ownershipFilename)

		require.Error(t, err)
		require.Contains(t, err.Error(), "owner must be specified when TESTTRACK_OWNERSHIP_FILE is defined")
	})

	t.Run("it errors out if owner file can't be found", func(t *testing.T) {
		var ownershipFilename = ".owners.yml"

		owner := "super_squad"
		err := validations.ValidateOwnerName(owner, ownershipFilename)

		require.Error(t, err)
		require.Contains(t, err.Error(), "open .owners.yml: no such file or directory")
	})

	t.Run("it errors out if owner can not be found in .squads.yml", func(t *testing.T) {
		var ownershipFilename = ".owners.yml"

		ownerContent := []byte("super_squad:\n  delayed_job_alert_slack_channel: '#super_squad'\n")
		os.WriteFile(ownershipFilename, ownerContent, 0644)

		owner := "not_super_squad"
		err := validations.ValidateOwnerName(owner, ownershipFilename)

		os.Remove(ownershipFilename)

		require.Error(t, err)
		require.Contains(t, err.Error(), "owner 'not_super_squad' is not defined in TESTTRACK_OWNERSHIP_FILE")
	})

	t.Run("it succeeds if the owner exists", func(t *testing.T) {
		var ownershipFilename = ".owners.yml"

		ownerContent := []byte("super_squad:\n  delayed_job_alert_slack_channel: '#super_squad'\n")
		os.WriteFile(ownershipFilename, ownerContent, 0644)

		owner := "super_squad"
		err := validations.ValidateOwnerName(owner, ownershipFilename)

		os.Remove(ownershipFilename)

		require.NoError(t, err)
	})

}

func StrPtr(value string) *string {
	return &value
}
