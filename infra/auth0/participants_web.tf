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
