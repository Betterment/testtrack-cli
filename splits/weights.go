package splits

import (
	"fmt"
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
