package fakeserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Betterment/testtrack-cli/fakeassignments"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/splits"
)

// v1Visitor is the JSON output type for V1 visitor endpoints
type v1Visitor struct {
	ID          string         `json:"id"`
	Assignments []v1Assignment `json:"assignments"`
}

// v4Visitor is the JSON output type for V4 visitor_config endpoints
type v4Visitor struct {
	ID          string         `json:"id"`
	Assignments []v4Assignment `json:"assignments"`
}

// v1Assignment is the JSON input/output type for V1 visitor endpoints
type v1Assignment struct {
	SplitName string `json:"split_name"`
	Variant   string `json:"variant"`
	Context   string `json:"context"`
	Unsynced  bool   `json:"unsynced"`
}

// v4Assignment is the JSON input/output type for V4 visitor_config endpoints
type v4Assignment struct {
	SplitName string `json:"split_name"`
	Variant   string `json:"variant"`
}

// v2AssignmentOverrideRequestBody is the JSON input for the V2 assignment override endpoint
type v2AssignmentOverrideRequestBody struct {
	Assignments []v1Assignment `json:"assignments"`
}

// v1VisitorConfig is the JSON output type for V1 visitor_config endpoints
type v1VisitorConfig struct {
	Splits  map[string]*splits.Weights `json:"splits"`
	Visitor v1Visitor                  `json:"visitor"`
}

// v2VisitorConfig is the JSON output type for V2 visitor_config endpoints
type v2VisitorConfig struct {
	Splits                   map[string]*v2Split `json:"splits"`
	Visitor                  v1Visitor           `json:"visitor"`
	ExperienceSamplingWeight int                 `json:"experience_sampling_weight"`
}

// v4VisitorConfig is the JSON output type for V4 visitor_config endpoints
type v4VisitorConfig struct {
	Splits                   []v4Split `json:"splits"`
	Visitor                  v4Visitor `json:"visitor"`
	ExperienceSamplingWeight int       `json:"experience_sampling_weight"`
}

// v2SplitRegistry is the JSON output type for V2 split_registry endpoint
type v2SplitRegistry struct {
	Splits                   map[string]*v2Split `json:"splits"`
	ExperienceSamplingWeight int                 `json:"experience_sampling_weight"`
}

// v4SplitRegistry is the JSON output type for V4 split_registry endpoint
type v4SplitRegistry struct {
	Splits                   []v4Split `json:"splits"`
	ExperienceSamplingWeight int       `json:"experience_sampling_weight"`
}

// v2Split is the JSON output type for V2 split_registry endpoint
type v2Split struct {
	Weights     map[string]int `json:"weights"`
	FeatureGate bool           `json:"feature_gate"`
}

// v4Split is the JSON output type for V4 split_registry endpoint
type v4Split struct {
	Name        string      `json:"name"`
	Variants    []v4Variant `json:"variants"`
	FeatureGate bool        `json:"feature_gate"`
}

// v4Split is the JSON output type for V4 split_registry endpoint
type v4Variant struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}

// v1SplitDetail is the JSON output type for the V1 split detail endpoint
type v1SplitDetail struct {
	Name               string            `json:"name"`
	Hypothesis         string            `json:"hypothesis"`
	AssignmentCriteria string            `json:"assignment_criteria"`
	Description        string            `json:"description"`
	Owner              string            `json:"owner"`
	Location           string            `json:"location"`
	Platform           string            `json:"platform"`
	VariantDetails     []v1VariantDetail `json:"variant_details"`
}

// v1VariantDetail is the JSON output type for variant details via the V1 split detail endpoint
type v1VariantDetail struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ScreenshotURL string `json:"screenshot_url"`
}

// v1AssignmentDetail is the JSON output type assignment details nested in the V1 visitor_detail endpoint
type v1AssignmentDetail struct {
	SplitLocation      string `json:"split_location"`
	SplitName          string `json:"split_name"`
	VariantName        string `json:"variant_name"`
	VariantDescription string `json:"variant_description"`
	AssignedAt         string `json:"assigned_at"`
}

func (s *server) routes() {
	s.handleGet(
		"/api/v1/split_registry",
		getV1SplitRegistry,
	)
	s.handleGet(
		"/api/v2/split_registry",
		getV2PlusSplitRegistry,
	)
	s.handlePostReturnNoContent(
		"/api/v1/assignment_event",
		postNoop,
	)
	s.handlePost(
		"/api/v1/identifier",
		postV1Identifier,
	)
	s.handlePost(
		"/api/v4/apps/{a}/versions/{v}/builds/{b}/identifier",
		postV4AppIdentifier,
	)
	s.handleGet(
		"/api/v1/visitors/{id}",
		getV1Visitor,
	)
	s.handleGet(
		"/api/v1/identifier_types/{t}/identifiers/{i}/visitor",
		getV1Visitor,
	)
	s.handleGet(
		"/api/v1/identifier_types/{t}/identifiers/{i}/visitor_detail",
		getV1VisitorDetail,
	)
	s.handlePostReturnNoContent(
		"/api/v1/assignment_override",
		postV1AssignmentOverride,
	)
	s.handlePostReturnNoContent(
		"/api/v2/visitors/{v}/assignment_overrides",
		postV2AssignmentOverride,
	)
	s.handleGet(
		"/api/v1/apps/{a}/versions/{v}/builds/{b}/visitors/{id}/config",
		getV1AppVisitorConfig,
	)
	s.handleGet(
		"/api/v4/apps/{a}/versions/{v}/builds/{b}/visitors/{id}/config",
		getV4AppVisitorConfig,
	)
	s.handleGet(
		"/api/v1/split_details/{id}",
		getV1SplitDetail,
	)
	s.handleGet(
		"/api/v3/builds/{b}/split_registry",
		getV2PlusSplitRegistry,
	)
	s.handleGet(
		"/api/v4/builds/{b}/split_registry",
		getV4SplitRegistry,
	)
}

func getV1SplitRegistry() (interface{}, error) {
	schema, err := schema.ReadMerged()
	if err != nil {
		return nil, err
	}
	splitRegistry := map[string]*splits.Weights{}
	for _, split := range schema.Splits {
		splitRegistry[split.Name], err = splits.NewWeights(split.Weights)
		if err != nil {
			return nil, err
		}
	}
	return splitRegistry, nil
}

func getV2PlusSplitRegistry() (interface{}, error) {
	schema, err := schema.ReadMerged()
	if err != nil {
		return nil, err
	}
	splitRegistry := map[string]*v2Split{}
	for _, split := range schema.Splits {
		isFeatureGate := splits.IsFeatureGateFromName(split.Name)
		weights, err := splits.NewWeights(split.Weights)
		if err != nil {
			return nil, err
		}
		splitRegistry[split.Name] = &v2Split{
			Weights:     *weights,
			FeatureGate: isFeatureGate,
		}
	}
	return v2SplitRegistry{
		Splits:                   splitRegistry,
		ExperienceSamplingWeight: 1,
	}, nil
}

func getV4SplitRegistry() (interface{}, error) {
	schema, err := schema.ReadMerged()
	if err != nil {
		return nil, err
	}
	v4Splits := make([]v4Split, 0, len(schema.Splits))
	for _, split := range schema.Splits {
		isFeatureGate := splits.IsFeatureGateFromName(split.Name)
		weights, err := splits.NewWeights(split.Weights)
		if err != nil {
			return nil, err
		}
		v4Variants := make([]v4Variant, 0, len(*weights))
		for variantName, weight := range *weights {
			v4Variants = append(v4Variants, v4Variant{
				Name:   variantName,
				Weight: weight,
			})
		}
		v4Splits = append(v4Splits, v4Split{
			Name:        split.Name,
			Variants:    v4Variants,
			FeatureGate: isFeatureGate,
		})
	}
	return v4SplitRegistry{
		Splits:                   v4Splits,
		ExperienceSamplingWeight: 1,
	}, nil
}

func postNoop(*http.Request) error {
	return nil
}

func postV1Identifier(*http.Request) (interface{}, error) {
	ivisitor, err := getV1Visitor()
	visitor := ivisitor.(v1Visitor)
	if err != nil {
		return nil, err
	}
	return map[string]v1Visitor{"visitor": visitor}, nil
}

func postV4AppIdentifier(*http.Request) (interface{}, error) {
	return getV4AppVisitorConfig()
}

func getV1Visitor() (interface{}, error) {
	assignments, err := fakeassignments.Read()
	if err != nil {
		return nil, err
	}
	v1Assignments := make([]v1Assignment, 0, len(*assignments))
	for split, variant := range *assignments {
		v1Assignments = append(v1Assignments, v1Assignment{
			SplitName: split,
			Variant:   variant,
			Context:   "fake_server",
			Unsynced:  false,
		})
	}
	return v1Visitor{
		ID:          "00000000-0000-0000-0000-000000000000",
		Assignments: v1Assignments,
	}, nil
}

func getV4Visitor() (interface{}, error) {
	assignments, err := fakeassignments.Read()
	if err != nil {
		return nil, err
	}
	v4Assignments := make([]v4Assignment, 0, len(*assignments))
	for split, variant := range *assignments {
		v4Assignments = append(v4Assignments, v4Assignment{
			SplitName: split,
			Variant:   variant,
		})
	}
	return v4Visitor{
		ID:          "00000000-0000-0000-0000-000000000000",
		Assignments: v4Assignments,
	}, nil
}

func getV1VisitorDetail() (interface{}, error) {
	assignments, err := fakeassignments.Read()
	if err != nil {
		return nil, err
	}
	v1AssignmentDetails := make([]v1AssignmentDetail, 0, len(*assignments))
	for split, variant := range *assignments {
		v1AssignmentDetails = append(v1AssignmentDetails, v1AssignmentDetail{
			SplitLocation:      "somewhere",
			SplitName:          split,
			VariantName:        variant,
			VariantDescription: "a very cool variant",
			AssignedAt:         "2019-05-02T16:57:36Z",
		})
	}
	return map[string][]v1AssignmentDetail{"assignment_details": v1AssignmentDetails}, nil
}

func postV1AssignmentOverride(r *http.Request) error {
	var assignment v1Assignment
	contentType := r.Header.Get("content-type")
	switch {
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		err := r.ParseForm()
		if err != nil {
			return err
		}
		assignment = v1Assignment{
			SplitName: r.PostForm.Get("split_name"),
			Variant:   r.PostForm.Get("variant"),
		}
	case strings.HasPrefix(contentType, "application/json"):
		requestBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(requestBytes, &assignment)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("got unexpected content type %s", contentType)
	}
	assignments, err := fakeassignments.Read()
	if err != nil {
		return err
	}

	(*assignments)[assignment.SplitName] = assignment.Variant
	err = fakeassignments.Write(assignments)
	if err != nil {
		return err
	}
	return nil
}

func postV2AssignmentOverride(r *http.Request) error {
	var assignments []v1Assignment
	contentType := r.Header.Get("content-type")
	switch {
	case strings.HasPrefix(contentType, "application/json"):
		requestBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		var assignmentBody v2AssignmentOverrideRequestBody
		err = json.Unmarshal(requestBytes, &assignmentBody)
		if err != nil {
			return err
		}
		assignments = assignmentBody.Assignments
	default:
		return fmt.Errorf("got unexpected content type %s", contentType)
	}
	storedAssignments, err := fakeassignments.Read()
	if err != nil {
		return err
	}

	for _, assignment := range assignments {
		(*storedAssignments)[assignment.SplitName] = assignment.Variant
	}
	err = fakeassignments.Write(storedAssignments)
	if err != nil {
		return err
	}
	return nil
}

func getV1AppVisitorConfig() (interface{}, error) {
	isplitRegistry, err := getV1SplitRegistry()
	splitRegistry := isplitRegistry.(map[string]*splits.Weights)
	if err != nil {
		return nil, err
	}
	ivisitor, err := getV1Visitor()
	visitor := ivisitor.(v1Visitor)
	if err != nil {
		return nil, err
	}
	return v1VisitorConfig{
		Splits:  splitRegistry,
		Visitor: visitor,
	}, nil
}

func getV4AppVisitorConfig() (interface{}, error) {
	isplitRegistry, err := getV4SplitRegistry()
	splitRegistry := isplitRegistry.(v4SplitRegistry)
	if err != nil {
		return nil, err
	}
	ivisitor, err := getV4Visitor()
	visitor := ivisitor.(v4Visitor)
	if err != nil {
		return nil, err
	}
	return v4VisitorConfig{
		Splits:                   splitRegistry.Splits,
		Visitor:                  visitor,
		ExperienceSamplingWeight: splitRegistry.ExperienceSamplingWeight,
	}, nil
}

func getV2AppVisitorConfig() (interface{}, error) {
	isplitRegistry, err := getV2PlusSplitRegistry()
	splitRegistry := isplitRegistry.(v2SplitRegistry)
	if err != nil {
		return nil, err
	}
	ivisitor, err := getV1Visitor()
	visitor := ivisitor.(v1Visitor)
	if err != nil {
		return nil, err
	}
	return v2VisitorConfig{
		Splits:                   splitRegistry.Splits,
		Visitor:                  visitor,
		ExperienceSamplingWeight: splitRegistry.ExperienceSamplingWeight,
	}, nil
}

func getV1SplitDetail() (interface{}, error) {
	return v1SplitDetail{
		Name:               "something",
		Hypothesis:         "my hypothesis",
		AssignmentCriteria: "assignment criteria go here...",
		Description:        "split description...",
		Owner:              "owner",
		Location:           "location",
		Platform:           "platform",
		VariantDetails: []v1VariantDetail{
			{
				Name:          "variant_a",
				Description:   "this is a fake description",
				ScreenshotURL: "https://example.org/a",
			},
			{
				Name:          "variant_b",
				Description:   "this is another fake description",
				ScreenshotURL: "https://example.org/b",
			},
		},
	}, nil
}
