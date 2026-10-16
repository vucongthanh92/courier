terraform {
  required_version = ">= 1.6"

  backend "s3" {
    bucket         = "courier-dev-tfstate-016170038458-ap-southeast-2"
    key            = "aws-dev/terraform.tfstate"
    region         = "ap-southeast-2"
    dynamodb_table = "courier-dev-terraform-lock"
    encrypt        = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = "dev"
      ManagedBy   = "terraform"
      Project     = var.project
    }
  }
}
