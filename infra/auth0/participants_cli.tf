resource "auth0_client" "participants_cli" {
  name            = "${var.api_name} (CLI)"
  app_type        = "native"
  grant_types     = ["urn:ietf:params:oauth:grant-type:device_code", "refresh_token"]
  oidc_conformant = true
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

resource "auth0_client_credentials" "participants_cli" {
  client_id             = auth0_client.participants_cli.id
  authentication_method = "none"
}
