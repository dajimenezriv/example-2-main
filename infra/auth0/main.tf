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

resource "auth0_client" "participants_web" {
  name                = "${var.api_name} (Web)"
  app_type            = "regular_web"
  grant_types         = ["authorization_code", "refresh_token"]
  callbacks           = ["${var.web_base_url}/callback"]
  allowed_logout_urls = ["${var.web_base_url}/"]
  oidc_conformant     = true

  jwt_configuration {
    alg = "RS256"
  }

  refresh_token {
    rotation_type = "rotating"
    # The two lifetimes from below are applied and the token can expire without invalidating it explicitely.
    expiration_type = "expiring"
    # The user will be logged out always after `token_lifetime`.
    token_lifetime = 2592000 # 30 days
    # If the token is not used for `idle_token_lifetime`, the token is expired.
    idle_token_lifetime          = 1296000 # 15 days
    infinite_token_lifetime      = false
    infinite_idle_token_lifetime = false
    # Multiple refresh token requests at the same time are not revoked. 10s as a grace period.
    leeway = 10
  }
}

resource "auth0_client_credentials" "participants_web" {
  client_id             = auth0_client.participants_web.id
  authentication_method = "client_secret_post"
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
