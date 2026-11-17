data "auth0_tenant" "current" {}

# The API this service protects. Tokens are RS256, verified via the tenant JWKS.
resource "auth0_resource_server" "participants" {
  name                                            = var.api_name
  identifier                                      = var.api_identifier
  signing_alg                                     = "RS256"
  token_lifetime                                  = var.token_lifetime
  allow_offline_access                            = false
  skip_consent_for_verifiable_first_party_clients = true

  # RBAC: user tokens only carry scopes the user has been granted.
  # Default dialect puts them in the "scope" claim, which internal/platform/auth reads.
  enforce_policies = true
  token_dialect    = "access_token"
}

resource "auth0_resource_server_scopes" "participants" {
  resource_server_identifier = auth0_resource_server.participants.identifier

  dynamic "scopes" {
    for_each = var.scopes
    content {
      name        = scopes.key
      description = scopes.value
    }
  }
}

# Machine-to-machine client for service callers, CI and local testing (client_credentials).
resource "auth0_client" "participants_m2m" {
  name        = "${var.api_name} (M2M)"
  app_type    = "non_interactive"
  grant_types = ["client_credentials"]

  jwt_configuration {
    alg = "RS256"
  }
}

resource "auth0_client_credentials" "participants_m2m" {
  client_id             = auth0_client.participants_m2m.id
  authentication_method = "client_secret_post"
}

resource "auth0_client_grant" "participants_m2m" {
  client_id = auth0_client.participants_m2m.id
  audience  = auth0_resource_server.participants.identifier
  scopes    = keys(var.scopes)

  depends_on = [auth0_resource_server_scopes.participants]
}
