locals {
  bucket_name = coalesce(var.bucket_name, "${var.name_prefix}-${var.folder_id}-uploads")
  image_url   = "cr.yandex/${yandex_container_registry.app.id}/${var.name_prefix}-api:${var.container_image_tag}"
  runtime_secret_entries = {
    JWT_SECRET                          = var.jwt_secret
    S3_ACCESS_KEY                       = yandex_iam_service_account_static_access_key.storage.access_key
    S3_SECRET_KEY                       = yandex_iam_service_account_static_access_key.storage.secret_key
    OAUTH_PROVIDER_GOOGLE_CLIENT_SECRET = var.oauth_google_client_secret
    OAUTH_PROVIDER_YANDEX_CLIENT_SECRET = var.oauth_yandex_client_secret
  }
  labels = {
    app        = "family-tree"
    managed_by = "terraform"
  }
}

resource "yandex_iam_service_account" "runtime" {
  name        = "${var.name_prefix}-runtime"
  description = "Runtime identity for the Family Tree serverless container."
}

resource "yandex_iam_service_account" "gateway" {
  name        = "${var.name_prefix}-gateway"
  description = "API Gateway identity for invoking the Family Tree container."
}

resource "yandex_iam_service_account" "storage" {
  name        = "${var.name_prefix}-storage"
  description = "Object Storage key owner for Family Tree uploads."
}

resource "yandex_resourcemanager_folder_iam_member" "runtime_ydb_editor" {
  folder_id = var.folder_id
  role      = "ydb.editor"
  member    = "serviceAccount:${yandex_iam_service_account.runtime.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "runtime_lockbox_payload_viewer" {
  folder_id = var.folder_id
  role      = "lockbox.payloadViewer"
  member    = "serviceAccount:${yandex_iam_service_account.runtime.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "runtime_registry_puller" {
  folder_id = var.folder_id
  role      = "container-registry.images.puller"
  member    = "serviceAccount:${yandex_iam_service_account.runtime.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "gateway_container_invoker" {
  folder_id = var.folder_id
  role      = "serverless-containers.containerInvoker"
  member    = "serviceAccount:${yandex_iam_service_account.gateway.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "storage_editor" {
  folder_id = var.folder_id
  role      = "storage.editor"
  member    = "serviceAccount:${yandex_iam_service_account.storage.id}"
}

resource "yandex_container_registry" "app" {
  name      = "${var.name_prefix}-registry"
  folder_id = var.folder_id
  labels    = local.labels
}

resource "yandex_container_repository" "api" {
  name = "${yandex_container_registry.app.id}/${var.name_prefix}-api"
}

resource "yandex_ydb_database_serverless" "app" {
  name                = "${var.name_prefix}-ydb"
  folder_id           = var.folder_id
  deletion_protection = var.ydb_deletion_protection
  labels              = local.labels
}

resource "yandex_iam_service_account_static_access_key" "storage" {
  service_account_id = yandex_iam_service_account.storage.id
  description        = "Family Tree Object Storage access key."

  depends_on = [yandex_resourcemanager_folder_iam_member.storage_editor]
}

resource "yandex_storage_bucket" "uploads" {
  access_key = yandex_iam_service_account_static_access_key.storage.access_key
  secret_key = yandex_iam_service_account_static_access_key.storage.secret_key
  bucket     = local.bucket_name

  depends_on = [yandex_resourcemanager_folder_iam_member.storage_editor]
}

resource "yandex_lockbox_secret" "runtime" {
  name        = "${var.name_prefix}-runtime-secrets"
  description = "Runtime secrets for the Family Tree API."
  folder_id   = var.folder_id
  labels      = local.labels
}

resource "yandex_lockbox_secret_version" "runtime" {
  secret_id = yandex_lockbox_secret.runtime.id

  dynamic "entries" {
    for_each = local.runtime_secret_entries
    content {
      key        = entries.key
      text_value = entries.value
    }
  }
}

resource "yandex_serverless_container" "api" {
  name               = "${var.name_prefix}-api"
  description        = "Family Tree API."
  folder_id          = var.folder_id
  service_account_id = yandex_iam_service_account.runtime.id
  memory             = var.container_memory
  cores              = var.container_cores
  core_fraction      = var.container_core_fraction
  execution_timeout  = var.container_execution_timeout
  concurrency        = var.container_concurrency
  labels             = local.labels

  runtime {
    type = "http"
  }

  image {
    url = local.image_url
    environment = {
      GIN_MODE                        = "release"
      LOG_LEVEL                       = "info"
      ENABLE_HSTS                     = "true"
      COOKIE_SECURE                   = "true"
      COOKIE_HTTP_ONLY                = "true"
      ALLOWED_ORIGINS                 = var.frontend_origin
      OAUTH_REDIRECT_BASE_URL         = var.frontend_origin
      OAUTH_ENABLED_PROVIDERS         = var.oauth_enabled_providers
      OAUTH_PROVIDER_GOOGLE_CLIENT_ID = var.oauth_google_client_id
      OAUTH_PROVIDER_YANDEX_CLIENT_ID = var.oauth_yandex_client_id
      ENABLE_MOCK_AUTH                = tostring(var.enable_mock_auth)
      SEED_TEST_DATA                  = tostring(var.seed_test_data)
      DATABASE_BACKEND                = "ydb"
      YDB_ENDPOINT                    = yandex_ydb_database_serverless.app.ydb_full_endpoint
      YDB_DATABASE                    = yandex_ydb_database_serverless.app.database_path
      YDB_AUTH_MODE                   = "metadata"
      S3_ENDPOINT                     = "https://storage.yandexcloud.net"
      S3_REGION                       = "ru-central1"
      S3_BUCKET                       = yandex_storage_bucket.uploads.bucket
      REDIS_URI                       = ""
    }
  }

  dynamic "secrets" {
    for_each = local.runtime_secret_entries
    content {
      id                   = yandex_lockbox_secret.runtime.id
      version_id           = yandex_lockbox_secret_version.runtime.id
      key                  = secrets.key
      environment_variable = secrets.key
    }
  }
}

resource "yandex_api_gateway" "api" {
  name        = "${var.name_prefix}-gateway"
  description = "Public API gateway for Family Tree."
  folder_id   = var.folder_id
  labels      = local.labels

  spec = templatefile("${path.module}/templates/api-gateway.yaml.tftpl", {
    title              = "Family Tree API"
    container_id       = yandex_serverless_container.api.id
    service_account_id = yandex_iam_service_account.gateway.id
  })
}
