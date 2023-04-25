package validations

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/Betterment/testtrack-cli/serializers"
	"gopkg.in/yaml.v2"
)

const appVersionMaxLength = 18 // This conforms to iOS version numering rules
const splitMaxLength = 128     // This is arbitrary but way bigger than you need and smaller than the column will fit

const defaultOwnershipFilename = "testtrack/owners.yml"

var prefixedSplitRegex = regexp.MustCompile(`^([a-z_\-\d]+)\.[a-z_\d]+$`)
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

// AutoPrefixAndValidateSplit automatically prefixes a split with app name
// according to flags and optionally validates it for presence in the schema
func AutoPrefixAndValidateSplit(paramName string, value *string, currentAppName string, schema *serializers.Schema, noPrefix, force bool) error {
	prefix := appNamePrefix(value)
	prefixed := fmt.Sprintf("%s.%s", currentAppName, *value)

	if noPrefix && prefix != nil {
		return fmt.Errorf("--no-prefix incompatible with prefix '%s'", *prefix)
	}

	if force {
		if noPrefix {
			return nil
		}

		if prefix != nil {
			return nil
		}
		*value = prefixed
		return nil
	}

	if !noPrefix {
		var candidate string

		if prefix == nil {
			candidate = prefixed
		} else {
			candidate = *value
		}

		err := SplitExistsInSchema(paramName, &candidate, schema)
		if err == nil {
			*value = candidate
			return nil
		}
	}

	return SplitExistsInSchema(paramName, value, schema)
}

// ValidateOwnerName ensures that if a .squads.yml file is present, the owner matches
// the list of owners in that file.
func ValidateOwnerName(owner string, ownershipFilename string) error {
	if ownershipFilename == "" {
		ownershipFilename = defaultOwnershipFilename
	}

	// If no ownership file exists, force owner to be empty. Otherwise pass validations.
	_, err := os.Stat(ownershipFilename)
	if os.IsNotExist(err) {
		if owner != "" {
			return fmt.Errorf("owner must be blank because ownership file (%s) could not be found", ownershipFilename)
		}

		return nil
	}

	// When the ownership file exists, owner must be specified and must be in the ownership file.
	if owner == "" {
		return fmt.Errorf("owner must be specified when ownership file (%s) exists", ownershipFilename)
	}

	fileBytes, err := ioutil.ReadFile(ownershipFilename)
	if err != nil {
		return err
	}

	ownersMap := make(map[string]*struct{})
	err = yaml.Unmarshal(fileBytes, ownersMap)
	if err != nil {
		return err
	}

	if !mapContainsValue(owner, ownersMap) {
		return fmt.Errorf("owner '%s' is not defined in ownership file (%s)", owner, ownershipFilename)
	}

	return nil
}

func mapContainsValue(value string, m map[string]*struct{}) bool {
	for key := range m {
		if key == value {
			return true
		}
	}
	return false
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

// ExperimentSuffix validates that an experiment name param ends in _experiment
func ExperimentSuffix(paramName string, value *string) error {
	if !strings.HasSuffix(*value, "_experiment") {
		return fmt.Errorf("%s '%s' must end in _experiment", paramName, *value)
	}
	return nil
}

// NonPrefixedExperiment validates that an experiment name param is valid with
// no app prefix
func NonPrefixedExperiment(paramName string, value *string) error {
	err := NonPrefixedSplit(paramName, value)
	if err != nil {
		return err
	}

	return ExperimentSuffix(paramName, value)
}

// FeatureGateSuffix validates that an experiment name param ends in _enabled
func FeatureGateSuffix(paramName string, value *string) error {
	if !strings.HasSuffix(*value, "_enabled") {
		return fmt.Errorf("%s '%s' must end in _enabled", paramName, *value)
	}
	return nil
}

// NonPrefixedFeatureGate validates that a feature_gate name param is valid
func NonPrefixedFeatureGate(paramName string, value *string) error {
	err := NonPrefixedSplit(paramName, value)
	if err != nil {
		return err
	}

	return FeatureGateSuffix(paramName, value)
}

// FeatureGate validates that a feature_gate name param is valid with no
// opinion on app prefix
func FeatureGate(paramName string, value *string) error {
	err := Split(paramName, value)
	if err != nil {
		return err
	}

	return FeatureGateSuffix(paramName, value)
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

// OptionalAppVersion validates that app version, if non-null, matches required format
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

// SplitExistsInSchema validates that a split exists in the schema
func SplitExistsInSchema(paramName string, value *string, schema *serializers.Schema) error {
	err := Presence(paramName, value)
	if err != nil {
		return err
	}
	for _, schemaSplit := range schema.Splits {
		if schemaSplit.Name == *value {
			return nil
		}
	}
	return fmt.Errorf("%s '%s' not found in schema", paramName, *value)
}

// VariantExistsInSchema validates that a variant exists for a split in the schema
func VariantExistsInSchema(paramName string, variant *string, split string, schema *serializers.Schema) error {
	err := Presence(paramName, variant)
	if err != nil {
		return err
	}
	for _, schemaSplit := range schema.Splits {
		if schemaSplit.Name == split {
			for _, item := range schemaSplit.Weights {
				v, ok := item.Key.(string)
				if !ok {
					return fmt.Errorf("variant %v is not a string", item.Key)
				}
				if v == *variant {
					return nil
				}
			}
			return fmt.Errorf("Split '%s' does not have variant '%s'", split, *variant)
		}
	}
	return fmt.Errorf("Split '%s' not found in schema", split)
}

func appNamePrefix(value *string) *string {
	matches := prefixedSplitRegex.FindStringSubmatch(*value)
	if matches != nil {
		return &matches[1]
	}
	return nil
}
