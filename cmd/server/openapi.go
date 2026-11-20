package main

import (
	"fmt"
	"os"

	"exercise/internal/platform/httpapi"

	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func printOpenAPI() error {
	api := humachi.New(chi.NewRouter(), httpapi.Config())
	registerOperations(api, setupApplication(nil))

	// Huma generates 3.1, but oapi-codegen only supports 3.0, so it is downgraded.
	spec, err := api.OpenAPI().DowngradeYAML()
	if err != nil {
		return fmt.Errorf("downgrade openapi spec: %w", err)
	}
	if _, err := os.Stdout.Write(spec); err != nil {
		return fmt.Errorf("write openapi spec: %w", err)
	}
	return nil
}
