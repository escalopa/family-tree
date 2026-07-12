variable "cloud_id" {
  description = "Yandex Cloud ID."
  type        = string
}

variable "yc_token" {
  description = "Yandex Cloud IAM token. Prefer passing it through TF_VAR_yc_token."
  type        = string
  sensitive   = true
  default     = null
}

variable "yc_service_account_key_file" {
  description = "Path to a Yandex Cloud service account key JSON file for CI/CD."
  type        = string
  default     = null
}

variable "folder_id" {
  description = "Yandex Cloud folder ID."
  type        = string
}

variable "zone" {
  description = "Default Yandex Cloud zone used by the provider."
  type        = string
  default     = "ru-central1-a"
}

variable "name_prefix" {
  description = "Prefix for created Yandex Cloud resources."
  type        = string
  default     = "family-tree"
}

variable "container_image_tag" {
  description = "Container image tag to deploy. Build and push this image before applying Terraform."
  type        = string
  default     = "latest"
}

variable "frontend_origin" {
  description = "Public frontend origin used for CORS and OAuth callbacks."
  type        = string
  default     = "https://family-tree.vercel.app"
}

variable "oauth_enabled_providers" {
  description = "Comma-separated OAuth providers for production."
  type        = string
  default     = "google"
}

variable "enable_mock_auth" {
  description = "Enable mock auth fallback. Production deploys should keep this false."
  type        = bool
  default     = false
}

variable "seed_test_data" {
  description = "Run test data seeding. Production deploys should keep this false."
  type        = bool
  default     = false
}

variable "oauth_google_client_id" {
  description = "Google OAuth client ID for production login."
  type        = string
}

variable "oauth_google_client_secret" {
  description = "Google OAuth client secret for production login."
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "JWT signing secret stored in Lockbox."
  type        = string
  sensitive   = true
}

variable "bucket_name" {
  description = "Object Storage bucket for member images. Defaults to a stable folder-scoped name."
  type        = string
  default     = null
}

variable "container_memory" {
  description = "Serverless container memory in MB."
  type        = number
  default     = 512
}

variable "container_cores" {
  description = "Serverless container cores."
  type        = number
  default     = 1
}

variable "container_core_fraction" {
  description = "Serverless container CPU core fraction."
  type        = number
  default     = 50
}

variable "container_execution_timeout" {
  description = "Serverless container execution timeout."
  type        = string
  default     = "30s"
}

variable "container_concurrency" {
  description = "Serverless container concurrency."
  type        = number
  default     = 8
}

variable "ydb_deletion_protection" {
  description = "Protect the YDB database from accidental deletion."
  type        = bool
  default     = true
}
