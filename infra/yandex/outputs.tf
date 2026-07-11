output "api_gateway_domain" {
  description = "Default API Gateway domain."
  value       = yandex_api_gateway.api.domain
}

output "api_gateway_url" {
  description = "Public API Gateway URL."
  value       = "https://${yandex_api_gateway.api.domain}"
}

output "serverless_container_url" {
  description = "Direct serverless container invoke URL."
  value       = yandex_serverless_container.api.url
}

output "container_image_url" {
  description = "Image URL Terraform expects to deploy."
  value       = local.image_url
}

output "container_registry_id" {
  description = "Container Registry ID."
  value       = yandex_container_registry.app.id
}

output "ydb_endpoint" {
  description = "YDB full endpoint for SDK clients."
  value       = yandex_ydb_database_serverless.app.ydb_full_endpoint
}

output "ydb_database" {
  description = "YDB database path."
  value       = yandex_ydb_database_serverless.app.database_path
}

output "uploads_bucket" {
  description = "Object Storage bucket for member uploads."
  value       = yandex_storage_bucket.uploads.bucket
}
