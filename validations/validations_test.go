package validations_test

import (
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

func StrPtr(value string) *string {
	return &value
}
