package main

import (
	"fmt"
	"io"

	"exercise/internal/platform/httpapi"

	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

// printOpenAPI writes the API spec as OpenAPI 3.0.3 YAML. Huma generates 3.1, but
// oapi-codegen only supports 3.0, so it is downgraded.
func printOpenAPI(w io.Writer) error {
	api := humachi.New(chi.NewRouter(), httpapi.Config())
	// Registering only reads handler metadata; nothing touches the (nil) DB pool.
	registerOperations(api, setupApplication(nil))

	spec, err := api.OpenAPI().DowngradeYAML()
	if err != nil {
		return fmt.Errorf("downgrade openapi spec: %w", err)
	}
	if _, err := w.Write(spec); err != nil {
		return fmt.Errorf("write openapi spec: %w", err)
	}
	return nil
}
