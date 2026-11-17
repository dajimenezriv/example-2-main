terraform {
  required_version = ">= 1.17"

  required_providers {
    auth0 = {
      source  = "auth0/auth0"
      version = "~> 1.0"
    }
  }
}

# Credentials come from env: AUTH0_DOMAIN, AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET
# (a bootstrap M2M app authorized for the Auth0 Management API).
provider "auth0" {}
