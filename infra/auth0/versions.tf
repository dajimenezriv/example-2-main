terraform {
  required_version = ">= 1.17"

  required_providers {
    auth0 = {
      source  = "auth0/auth0"
      version = "~> 1.0"
    }
  }
}

# Runs as a bootstrap M2M app authorized for the Auth0 Management API.
provider "auth0" {
  domain        = var.auth0_domain
  client_id     = var.management_client_id
  client_secret = var.management_client_secret
}
