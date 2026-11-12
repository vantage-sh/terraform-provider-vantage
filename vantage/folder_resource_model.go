package vantage

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	modelsv2 "github.com/vantage-sh/vantage-go/vantagev2/models"
	foldersv2 "github.com/vantage-sh/vantage-go/vantagev2/vantage/folders"
)

// The generated SDK marks saved_filter_tokens omitempty, so even a non-nil
// empty slice disappears from JSON. Always include the array to preserve the
// ability to clear saved filters while keeping the generated client untouched.
func updateFolderWithSavedFilterTokens(model *modelsv2.UpdateFolder) foldersv2.ClientOption {
	return func(operation *runtime.ClientOperation) {
		original := operation.Params
		operation.Params = runtime.ClientRequestWriterFunc(func(req runtime.ClientRequest, registry strfmt.Registry) error {
			if err := original.WriteToRequest(req, registry); err != nil {
				return err
			}
			return req.SetBodyParam(struct {
				*modelsv2.UpdateFolder
				SavedFilterTokens []string `json:"saved_filter_tokens"`
			}{model, model.SavedFilterTokens})
		})
	}
}
