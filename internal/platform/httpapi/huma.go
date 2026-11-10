package httpapi

import "github.com/danielgtaylor/huma/v2"

func Config() huma.Config {
	cfg := huma.DefaultConfig("Participants API", "1.0.0")
	// Drop the `$schema` field in response bodies.
	cfg.CreateHooks = nil
	return cfg
}

func Error500InternalServerError(err error) huma.StatusError {
	// Log err
	return huma.Error500InternalServerError("internal server error")
}
