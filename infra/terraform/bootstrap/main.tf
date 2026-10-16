locals {
  state_bucket_name = "${var.project}-dev-tfstate-${var.account_id}-${var.aws_region}"
}

resource "aws_s3_bucket" "terraform_state" {
  bucket = local.state_bucket_name

  tags = {
    Environment = "dev"
    ManagedBy   = "terraform"
    Purpose     = "terraform-state"
    Project     = var.project
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket                  = aws_s3_bucket.terraform_state.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_dynamodb_table" "terraform_lock" {
  name         = "${var.project}-dev-terraform-lock"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = {
    Environment = "dev"
    ManagedBy   = "terraform"
    Purpose     = "terraform-lock"
    Project     = var.project
  }
}
