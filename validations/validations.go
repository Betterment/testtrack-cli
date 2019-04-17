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

// Split validates that a `split` param is valid
func Split(splitName *string) error {
	splitParam := "split_name"
	return SnakeCaseParam(splitName, &splitParam)
}

// FeatureGate validates that a `feature_gate` param is valid
func FeatureGate(featureGateName *string) error {
	featureGateParam := "feature_gate_name"
	err := SnakeCaseParam(featureGateName, &featureGateParam)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(*featureGateName, "_enabled") {
		return fmt.Errorf("feature_gate_name '%s' must end in _enabled", *featureGateName)
	}
	return nil
}

// Presence validates that a a param is present
func Presence(value, paramName *string) error {
	if value == nil || len(*value) == 0 {
		return fmt.Errorf("%s must be present", *paramName)
	}
	return nil
}

// OptionalSnakeCaseParam validates that a param is snake case alphanumeric if present
func OptionalSnakeCaseParam(name, paramName *string) error {
	if name != nil {
		return SnakeCaseParam(name, paramName)
	}
	return nil
}

// SnakeCaseParam validates that a param is snake case alphanumeric
func SnakeCaseParam(name, paramName *string) error {
	err := Presence(name, paramName)
	if err != nil {
		return err
	}

	if !snakeCaseRegex.MatchString(*name) {
		return fmt.Errorf("%s '%s' must be snake_case alphanumeric", *paramName, *name)
	}
	return nil
}

// OptionalAppVersion validates that a param, if non-null, matches required format
func OptionalAppVersion(version, paramName *string) error {
	if version != nil {
		if !appVersionRegex.MatchString(*version) {
			return fmt.Errorf("%s '%s' must be made up of no more than three integers with dots in between", *paramName, *version)
		}

		if len(*version) > appVersionMaxLength {
			return fmt.Errorf("%s '%s' must be %d characters or less", *paramName, *version, appVersionMaxLength)
		}
	}
	return nil
}
