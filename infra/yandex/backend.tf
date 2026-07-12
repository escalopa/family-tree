terraform {
  backend "s3" {
    bucket = "escalopa-tfstate"
    key    = "family-tree/prod/terraform.tfstate"
    region = "ru-central1"

    endpoint                    = "https://storage.yandexcloud.net"
    force_path_style            = true
    skip_credentials_validation = true
    skip_metadata_api_check     = true
    skip_region_validation      = true
    skip_requesting_account_id  = true
    skip_s3_checksum            = true
  }
}
