output "tf_state_bucket" {
  description = "Object Storage bucket for Terraform production state."
  value       = yandex_storage_bucket.terraform_state.bucket
}

output "tf_state_access_key_id" {
  description = "Access key ID for the Terraform S3 backend."
  value       = yandex_iam_service_account_static_access_key.terraform_state.access_key
  sensitive   = true
}

output "tf_state_secret_access_key" {
  description = "Secret access key for the Terraform S3 backend."
  value       = yandex_iam_service_account_static_access_key.terraform_state.secret_key
  sensitive   = true
}

output "yc_service_account_key_json" {
  description = "Service account authorized key JSON for the Yandex Terraform provider and Container Registry login."
  value = jsonencode({
    id                 = yandex_iam_service_account_key.github_actions.id
    service_account_id = yandex_iam_service_account.github_actions.id
    created_at         = yandex_iam_service_account_key.github_actions.created_at
    key_algorithm      = "RSA_2048"
    public_key         = yandex_iam_service_account_key.github_actions.public_key
    private_key        = yandex_iam_service_account_key.github_actions.private_key
  })
  sensitive = true
}
