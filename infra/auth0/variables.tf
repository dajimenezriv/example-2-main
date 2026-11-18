variable "api_identifier" {
  description = "Auth0 API identifier. Becomes the token audience (AUTH_AUDIENCE)."
  type        = string
  default     = "https://participants-api"
}

variable "api_name" {
  description = "Display name of the API in Auth0."
  type        = string
  default     = "Participants API"
}

variable "token_lifetime" {
  description = "Access token lifetime in seconds."
  type        = number
  default     = 3600
}

variable "auth0_domain" {
  description = "Auth0 tenant domain."
  type        = string
}

variable "management_client_id" {
  description = "Client ID of the bootstrap Management API app the provider runs as."
  type        = string
}

variable "management_client_secret" {
  description = "Client secret of the bootstrap Management API app."
  type        = string
  sensitive   = true
}

variable "admin_email" {
  description = "Email of the bootstrap admin (passwordless login). Set via TF_VAR_admin_email in infra/auth0/.env."
  type        = string
}

variable "scopes" {
  description = "Permissions exposed by the API (scope value => description)."
  type        = map(string)
  default = {
    "read:actors"    = "Read actors"
    "write:actors"   = "Create and update actors"
    "read:accounts"  = "Read accounts"
    "write:accounts" = "Create accounts"
  }
}

variable "web_base_url" {
  description = "Web base URL."
  type        = string
  default     = "http://localhost:8080"
}
