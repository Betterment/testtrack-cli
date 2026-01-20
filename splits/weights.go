package splits

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v2"
)

// Weights represents the weightings of a split
type Weights map[string]int

// NewWeights creates a Weights instance from a map, validating that weights sum to 100
func NewWeights(weights map[string]int) (*Weights, error) {
	cumulativeWeight := 0
	for _, weight := range weights {
		if weight < 0 {
			return nil, fmt.Errorf("weight %d is less than zero", weight)
		}
		cumulativeWeight += weight
	}
	if cumulativeWeight != 100 {
		return nil, fmt.Errorf("weights must sum to 100, got %d", cumulativeWeight)
	}
	w := Weights(weights)
	return &w, nil
}

// WeightsFromYAML converts YAML-serializable weights to a weights map
func WeightsFromYAML(yamlWeights yaml.MapSlice) (*Weights, error) {
	weights := make(map[string]int)
	for _, item := range yamlWeights {
		variant, ok := item.Key.(string)
		if !ok {
			return nil, fmt.Errorf("variant %v is not a string", item.Key)
		}
		weight, ok := item.Value.(int)
		if !ok {
			return nil, fmt.Errorf("weighting %v is not an int", item.Value)
		}
		weights[variant] = weight
	}
	return NewWeights(weights)
}

// ToYAML converts weights to a YAML-serializable representation
func (w *Weights) ToYAML() yaml.MapSlice {
	var variants = make([]string, 0, len(*w))
	for variant := range *w {
		variants = append(variants, variant)
	}
	sort.Strings(variants)
	weightsYaml := make(yaml.MapSlice, 0, len(variants))
	for _, variant := range variants {
		weightsYaml = append(weightsYaml, yaml.MapItem{Key: variant, Value: (*w)[variant]})
	}
	return weightsYaml
}

// Merge newWeights over weights
func (w *Weights) Merge(newWeights Weights) {
	for variant := range *w {
		(*w)[variant] = 0
	}
	for variant, weight := range newWeights {
		(*w)[variant] = weight
	}
}

// ReweightToDecision sets weights to 100% one variant
func (w *Weights) ReweightToDecision(variant string) error {
	if _, ok := (*w)[variant]; !ok {
		return fmt.Errorf("couldn't locate variant %s", variant)
	}
	w.Merge(Weights{variant: 100})
	return nil
}
