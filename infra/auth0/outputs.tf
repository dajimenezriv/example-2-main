output "auth_issuer" {
  description = "AUTH_ISSUER for the service (trailing slash matters)."
  value       = "https://${var.auth0_domain}/"
}

output "auth_audience" {
  description = "AUTH_AUDIENCE for the service."
  value       = auth0_resource_server.participants.identifier
}

output "m2m_client_id" {
  value = auth0_client.participants_m2m.client_id
}

output "m2m_client_secret" {
  value     = auth0_client_credentials.participants_m2m.client_secret
  sensitive = true
}

output "cli_client_id" {
  value = auth0_client.participants_cli.client_id
}
