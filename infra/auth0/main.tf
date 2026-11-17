data "auth0_tenant" "current" {}

# The API this service protects. Tokens are RS256, verified via the tenant JWKS.
resource "auth0_resource_server" "participants" {
  name                                            = var.api_name
  identifier                                      = var.api_identifier
  signing_alg                                     = "RS256"
  token_lifetime                                  = var.token_lifetime
  allow_offline_access                            = true
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

resource "auth0_connection" "google" {
  name     = "google-oauth2"
  strategy = "google-oauth2"

  options {
    scopes = ["email", "profile"]
  }
}

resource "auth0_connection_clients" "google" {
  connection_id   = auth0_connection.google.id
  enabled_clients = [auth0_client.participants_web.id]
}
