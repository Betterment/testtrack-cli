package validations

import (
	"fmt"
	"regexp"
	"strings"
)

const appVersionMaxLength = 18 // This conforms to iOS version numering rules
const splitMaxLength = 128     // This is arbitrary but way bigger than you need and smaller than the column will fit

var snakeCaseRegex = regexp.MustCompile(`^[a-z_\d]+$`)
var decimalIntegerRegexPart = `(?:0|[1-9]\d*)`
var appVersionRegex = regexp.MustCompile(strings.Join([]string{
	`^(?:`,
	decimalIntegerRegexPart,
	`\.){0,2}`,
	decimalIntegerRegexPart,
	`$`,
}, ""))

// SplitName validates that a `split_name` param is valid
func SplitName(splitName *string) error {
	return SnakeCaseParam(splitName, "split_name")
}

// FeatureGateName validates that a `feature_gate_name` param is valid
func FeatureGateName(featureGateName *string) error {
	err := SnakeCaseParam(featureGateName, "feature_gate_name")
	if err != nil {
		return err
	}

	if !strings.HasSuffix(*featureGateName, "_enabled") {
		return fmt.Errorf("feature_gate_name '%s' must end in _enabled", *featureGateName)
	}
	return nil
}

// Presence validates that a a param is present
func Presence(value *string, paramName string) error {
	if value == nil || len(*value) == 0 {
		return fmt.Errorf("%s must be present", paramName)
	}
	return nil
}

// OptionalSnakeCaseParam validates that a param is snake case alphanumeric if present
func OptionalSnakeCaseParam(name *string, paramName string) error {
	if name != nil {
		return SnakeCaseParam(name, paramName)
	}
	return nil
}

// SnakeCaseParam validates that a param is snake case alphanumeric
func SnakeCaseParam(name *string, paramName string) error {
	err := Presence(name, paramName)
	if err != nil {
		return err
	}

	if !snakeCaseRegex.MatchString(*name) {
		return fmt.Errorf("%s '%s' must be snake_case alphanumeric", paramName, *name)
	}
	return nil
}

// OptionalAppVersion validates that a param, if non-null, matches required format
func OptionalAppVersion(version *string, paramName string) error {
	if version != nil {
		if !appVersionRegex.MatchString(*version) {
			return fmt.Errorf("%s '%s' must be made up of no more than three integers with dots in between", paramName, *version)
		}

		if len(*version) > appVersionMaxLength {
			return fmt.Errorf("%s '%s' must be %d characters or less", paramName, *version, appVersionMaxLength)
		}
	}
	return nil
}
