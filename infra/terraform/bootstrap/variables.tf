variable "aws_region" {
  type    = string
  default = "ap-southeast-2"
}

variable "account_id" {
  description = "AWS account ID, used to make the state bucket globally unique."
  type        = string
  default     = "016170038458"
}

variable "project" {
  type    = string
  default = "courier"
}
