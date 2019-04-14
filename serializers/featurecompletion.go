package serializers

import "fmt"

// SerializerVersion is the current version of the migration file format so we can evolve over time
const SerializerVersion = 1

// Migration is the YAML-marshalable root of a migration file
type Migration struct {
	SerializerVersion int                `yaml:"serializer_version"`
	FeatureCompletion *FeatureCompletion `yaml:"feature_completion,omitempty"`
}

// FeatureCompletion is the YAML-marshalable representation of a FeatureCompletion
type FeatureCompletion struct {
	FeatureGate string  `yaml:"feature_gate"`
	Version     *string `yaml:"version"`
}

func main() {
	fmt.Println("vim-go")
}
