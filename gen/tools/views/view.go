// Code generated by goa v3.16.1, DO NOT EDIT.
//
// tools views
//
// Command:
// $ goa gen github.com/arduino/arduino-create-agent/design

package views

import (
	goa "goa.design/goa/v3/pkg"
)

// ToolCollection is the viewed result type that is projected based on a view.
type ToolCollection struct {
	// Type to project
	Projected ToolCollectionView
	// View to render
	View string
}

// Operation is the viewed result type that is projected based on a view.
type Operation struct {
	// Type to project
	Projected *OperationView
	// View to render
	View string
}

// ToolCollectionView is a type that runs validations on a projected type.
type ToolCollectionView []*ToolView

// ToolView is a type that runs validations on a projected type.
type ToolView struct {
	// The name of the tool
	Name *string
	// The version of the tool
	Version *string
	// The packager of the tool
	Packager *string
}

// OperationView is a type that runs validations on a projected type.
type OperationView struct {
	// The status of the operation
	Status *string
}

var (
	// ToolCollectionMap is a map indexing the attribute names of ToolCollection by
	// view name.
	ToolCollectionMap = map[string][]string{
		"default": {
			"name",
			"version",
			"packager",
		},
	}
	// OperationMap is a map indexing the attribute names of Operation by view name.
	OperationMap = map[string][]string{
		"default": {
			"status",
		},
	}
	// ToolMap is a map indexing the attribute names of Tool by view name.
	ToolMap = map[string][]string{
		"default": {
			"name",
			"version",
			"packager",
		},
	}
)

// ValidateToolCollection runs the validations defined on the viewed result
// type ToolCollection.
func ValidateToolCollection(result ToolCollection) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateToolCollectionView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidateOperation runs the validations defined on the viewed result type
// Operation.
func ValidateOperation(result *Operation) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateOperationView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidateToolCollectionView runs the validations defined on
// ToolCollectionView using the "default" view.
func ValidateToolCollectionView(result ToolCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidateToolView(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateToolView runs the validations defined on ToolView using the
// "default" view.
func ValidateToolView(result *ToolView) (err error) {
	if result.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "result"))
	}
	if result.Version == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("version", "result"))
	}
	if result.Packager == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("packager", "result"))
	}
	return
}

// ValidateOperationView runs the validations defined on OperationView using
// the "default" view.
func ValidateOperationView(result *OperationView) (err error) {
	if result.Status == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("status", "result"))
	}
	return
}
