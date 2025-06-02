package validations_test

import (
	"os"
	"path/filepath"
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
				{
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
				{
					Name: "foo_enabled",
				},
				{
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
				{
					Name: "my_app.foo_enabled",
				},
				{
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
	t.Run("it succeeds with no owner if ownershipFilename is undefined and the default file does not exist", func(t *testing.T) {
		err := validations.ValidateOwnerName("")
		require.NoError(t, err)
	})

	t.Run("it fails with an owner if ownershipFilename is undefined and the default file does not exist ", func(t *testing.T) {
		err := validations.ValidateOwnerName("super_owner")
		require.Error(t, err)
		require.Contains(t, err.Error(), "owner must be blank because ownership file (testtrack/owners.yml) could not be found")
	})

	t.Run("it fails if using default ownership file and owner is blank", func(t *testing.T) {
		WriteOwnershipFile(validations.DefaultOwnershipFilePath)

		err := validations.ValidateOwnerName("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "owner must be specified when ownership file (testtrack/owners.yml) exists")

		RemoveOwnershipFile(validations.DefaultOwnershipFilePath)
	})

	t.Run("it fails if using specified ownership file and owner is blank", func(t *testing.T) {
		WriteOwnershipFile(".owners.yml")
		t.Setenv("TESTTRACK_OWNERSHIP_FILE", ".owners.yml")

		err := validations.ValidateOwnerName("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "owner must be specified when ownership file (.owners.yml) exists")

		RemoveOwnershipFile(".owners.yml")
	})

	t.Run("it succeeds if using default ownership file and owner exists", func(t *testing.T) {
		WriteOwnershipFile(validations.DefaultOwnershipFilePath)

		err := validations.ValidateOwnerName("super_owner")
		require.NoError(t, err)

		RemoveOwnershipFile(validations.DefaultOwnershipFilePath)
	})

	t.Run("it succeeds if using specified ownership file and owner exists", func(t *testing.T) {
		WriteOwnershipFile(".owners.yml")
		t.Setenv("TESTTRACK_OWNERSHIP_FILE", ".owners.yml")

		err := validations.ValidateOwnerName("super_owner")
		require.NoError(t, err)

		RemoveOwnershipFile(".owners.yml")
	})

	t.Run("it fails if using default ownership file and owner does not exist", func(t *testing.T) {
		WriteOwnershipFile(validations.DefaultOwnershipFilePath)

		err := validations.ValidateOwnerName("superb_owner")
		require.Error(t, err)
		require.Contains(t, err.Error(), "owner 'superb_owner' is not defined in ownership file (testtrack/owners.yml)")

		RemoveOwnershipFile(validations.DefaultOwnershipFilePath)
	})

	t.Run("it fails if using specified ownership file and owner does not exist", func(t *testing.T) {
		WriteOwnershipFile(".owners.yml")
		t.Setenv("TESTTRACK_OWNERSHIP_FILE", ".owners.yml")

		err := validations.ValidateOwnerName("superb_owner")
		require.Error(t, err)
		require.Contains(t, err.Error(), "owner 'superb_owner' is not defined in ownership file (.owners.yml)")

		RemoveOwnershipFile(".owners.yml")
	})
}

func StrPtr(value string) *string {
	return &value
}

func WriteOwnershipFile(ownershipFilename string) {
	if _, err := os.Stat(ownershipFilename); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(ownershipFilename), 0700)
	}

	ownerContent := []byte("super_owner:\n  delayed_job_alert_slack_channel: '#super_owner'\n")
	os.WriteFile(ownershipFilename, ownerContent, 0644)
}

func RemoveOwnershipFile(ownershipFilename string) {
	os.Remove(ownershipFilename)
	os.RemoveAll(filepath.Dir(ownershipFilename))
}
