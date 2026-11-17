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
