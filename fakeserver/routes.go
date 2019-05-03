package fakeserver

import (
	"encoding/json"

	"github.com/Betterment/testtrack-cli/fakeassignments"
	"github.com/Betterment/testtrack-cli/schema"
	"github.com/Betterment/testtrack-cli/splits"
)

// v1Visitor is the JSON output type for V1 visitor endpoints
type v1Visitor struct {
	ID          string         `json:"id"`
	Assignments []v1Assignment `json:"assignments"`
}

// v1Assignment is the JSON input/output type for V1 visitor endpoints
type v1Assignment struct {
	SplitName string `json:"split_name"`
	Variant   string `json:"variant"`
	Context   string `json:"context"`
	Unsynced  bool   `json:"unsynced"`
}

// v1VisitorConfig is the JSON output type for V1 visitor_config endpoints
type v1VisitorConfig struct {
	Splits  interface{} `json:"splits"`
	Visitor interface{} `json:"visitor"`
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
	s.handlePost(
		"/api/v1/assignment_event",
		postNoop,
	)
	s.handlePost(
		"/api/v1/identifier",
		postNoop,
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
	s.handlePost(
		"/api/v1/assignment_override",
		postV1AssignmentOverride,
	)
	s.handleGet(
		"/api/v1/apps/{a}/versions/{v}/builds/{b}/visitors/{id}/config",
		getV1AppVisitorConfig,
	)
	s.handleGet(
		"/api/v1/apps/{a}/versions/{v}/builds/{b}/identifier_types/{t}/identifiers/{i}/visitor_config",
		getV1AppVisitorConfig,
	)
	s.handleGet(
		"/api/v1/split_details/{id}",
		getV1SplitDetail,
	)
}

func getV1SplitRegistry() (interface{}, error) {
	schema, err := schema.Read()
	if err != nil {
		return nil, err
	}
	splitRegistry := map[string]*splits.Weights{}
	for _, split := range schema.Splits {
		splitRegistry[split.Name], err = splits.WeightsFromYAML(split.Weights)
		if err != nil {
			return nil, err
		}
	}
	return splitRegistry, nil
}

func postNoop([]byte) error {
	return nil
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

func postV1AssignmentOverride(requestBytes []byte) error {
	var assignment v1Assignment
	err := json.Unmarshal(requestBytes, &assignment)
	if err != nil {
		return err
	}
	assignments, err := fakeassignments.Read()
	(*assignments)[assignment.SplitName] = assignment.Variant
	err = fakeassignments.Write(assignments)
	if err != nil {
		return err
	}
	return nil
}

func getV1AppVisitorConfig() (interface{}, error) {
	splitRegistry, err := getV1SplitRegistry()
	if err != nil {
		return nil, err
	}
	visitor, err := getV1Visitor()
	if err != nil {
		return nil, err
	}
	return v1VisitorConfig{
		Splits:  splitRegistry,
		Visitor: visitor,
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
			v1VariantDetail{
				Name:          "variant_a",
				Description:   "this is a fake description",
				ScreenshotURL: "https://example.org/a",
			},
			v1VariantDetail{
				Name:          "variant_b",
				Description:   "this is another fake description",
				ScreenshotURL: "https://example.org/b",
			},
		},
	}, nil
}
