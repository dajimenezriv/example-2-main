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
