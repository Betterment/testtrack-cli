package validations

import (
	"fmt"
	"regexp"
	"strings"
)

const appVersionMaxLength = 18 // This conforms to iOS version numering rules
const splitMaxLength = 128     // This is arbitrary but way bigger than you need and smaller than the column will fit

var prefixedSplitRegex = regexp.MustCompile(`^[a-z_\-\d]+\.[a-z_\d]+$`)
var nonPrefixedSplitRegex = regexp.MustCompile(`^[a-z_\d]+$`)
var ambiPrefixedSplitRegex = regexp.MustCompile(`^(?:[a-z_\-\d]+\.)?[a-z_\d]+$`)
var snakeCaseRegex = regexp.MustCompile(`^[a-z_\d]+$`)
var decimalIntegerRegexPart = `(?:0|[1-9]\d*)`
var appVersionRegex = regexp.MustCompile(strings.Join([]string{
	`^(?:`,
	decimalIntegerRegexPart,
	`\.){0,2}`,
	decimalIntegerRegexPart,
	`$`,
}, ""))

// PrefixedSplit validates that split name param is valid with an app prefix
func PrefixedSplit(paramName string, value *string) error {
	err := Presence(paramName, value)
	if err != nil {
		return err
	}

	if !prefixedSplitRegex.MatchString(*value) {
		return fmt.Errorf("%s '%s' must be an app-name prefixed split name", paramName, *value)
	}
	return nil
}

// NonPrefixedSplit validates that a split name param is valid with no app prefix
func NonPrefixedSplit(paramName string, value *string) error {
	err := Presence(paramName, value)
	if err != nil {
		return err
	}

	if !nonPrefixedSplitRegex.MatchString(*value) {
		return fmt.Errorf("%s '%s' must be snake_case alphanumeric with no app prefix", paramName, *value)
	}
	return nil
}

// Split validates that a split name param is valid with no opinion on app prefix
func Split(paramName string, value *string) error {
	err := Presence(paramName, value)
	if err != nil {
		return err
	}

	if !ambiPrefixedSplitRegex.MatchString(*value) {
		return fmt.Errorf("%s '%s' must be a valid split name", paramName, *value)
	}
	return nil
}

// NonPrefixedFeatureGate validates that a `feature_gate_name` param is valid
func NonPrefixedFeatureGate(paramName string, value *string) error {
	err := NonPrefixedSplit(paramName, value)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(*value, "_enabled") {
		return fmt.Errorf("%s '%s' must end in _enabled", paramName, *value)
	}
	return nil
}

// FeatureGate validates that a feature_gate name param is valid with no
// opinion on app prefix
func FeatureGate(paramName string, value *string) error {
	err := Split(paramName, value)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(*value, "_enabled") {
		return fmt.Errorf("%s '%s' must end in _enabled", paramName, *value)
	}
	return nil
}

// Presence validates that a param is present
func Presence(paramName string, value *string) error {
	if value == nil || len(*value) == 0 {
		return fmt.Errorf("%s must be present", paramName)
	}
	return nil
}

// OptionalSnakeCaseParam validates that a param is snake case alphanumeric with potential dots if present
func OptionalSnakeCaseParam(paramName string, value *string) error {
	if value != nil && len(*value) > 0 {
		return SnakeCaseParam(paramName, value)
	}
	return nil
}

// SnakeCaseParam validates that a param is snake case alphanumeric
func SnakeCaseParam(paramName string, value *string) error {
	err := Presence(paramName, value)
	if err != nil {
		return err
	}

	if !snakeCaseRegex.MatchString(*value) {
		return fmt.Errorf("%s '%s' must be snake_case alphanumeric", paramName, *value)
	}
	return nil
}

// OptionalAppVersion validates that a param, if non-null, matches required format
func OptionalAppVersion(paramName string, value *string) error {
	if value != nil && len(*value) > 0 {
		if !appVersionRegex.MatchString(*value) {
			return fmt.Errorf("%s '%s' must be made up of no more than three integers with dots in between", paramName, *value)
		}

		if len(*value) > appVersionMaxLength {
			return fmt.Errorf("%s '%s' must be %d characters or less", paramName, *value, appVersionMaxLength)
		}
	}
	return nil
}
