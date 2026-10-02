terraform {
  required_version = ">= 1.4"

  required_providers {
    googleworkspace = {
      source  = "macadmins/googleworkspace"
      version = ">= 1.5.0"
    }
  }
}
