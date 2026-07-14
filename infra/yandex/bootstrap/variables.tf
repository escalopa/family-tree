variable "cloud_id" {
  description = "Yandex Cloud ID."
  type        = string
}

variable "folder_id" {
  description = "Yandex Cloud folder ID."
  type        = string
}

variable "zone" {
  description = "Default Yandex Cloud zone."
  type        = string
  default     = "ru-central1-a"
}

variable "yc_token" {
  description = "Yandex Cloud IAM token used for the bootstrap apply."
  type        = string
  sensitive   = true
}

variable "name_prefix" {
  description = "Prefix for bootstrap resources."
  type        = string
  default     = "family-tree"
}

variable "tf_state_bucket_name" {
  description = "Object Storage bucket used by Terraform remote state."
  type        = string
  default     = "family-tree-b1gkimk9k36atshi4uto-tfstate"
}
