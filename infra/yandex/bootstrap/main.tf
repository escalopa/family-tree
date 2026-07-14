locals {
  labels = {
    app        = "family-tree"
    managed_by = "terraform"
    purpose    = "ci-bootstrap"
  }
}

resource "yandex_iam_service_account" "github_actions" {
  name        = "${var.name_prefix}-github-actions"
  description = "GitHub Actions deploy identity for Family Tree production."
}

resource "yandex_resourcemanager_folder_iam_member" "github_actions_admin" {
  folder_id = var.folder_id
  role      = "admin"
  member    = "serviceAccount:${yandex_iam_service_account.github_actions.id}"
}

resource "yandex_iam_service_account_key" "github_actions" {
  service_account_id = yandex_iam_service_account.github_actions.id
  description        = "GitHub Actions Terraform provider key."
  key_algorithm      = "RSA_2048"

  depends_on = [yandex_resourcemanager_folder_iam_member.github_actions_admin]
}

resource "yandex_iam_service_account_static_access_key" "terraform_state" {
  service_account_id = yandex_iam_service_account.github_actions.id
  description        = "GitHub Actions Terraform state backend access key."

  depends_on = [yandex_resourcemanager_folder_iam_member.github_actions_admin]
}

resource "yandex_storage_bucket" "terraform_state" {
  bucket     = var.tf_state_bucket_name
  folder_id  = var.folder_id
  access_key = yandex_iam_service_account_static_access_key.terraform_state.access_key
  secret_key = yandex_iam_service_account_static_access_key.terraform_state.secret_key
  tags       = local.labels

  versioning {
    enabled = true
  }
}
