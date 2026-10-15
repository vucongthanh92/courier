variable "aws_region" {
  type    = string
  default = "ap-southeast-1"
}

variable "project" {
  type    = string
  default = "courier"
}

variable "vpc_cidr" {
  type    = string
  default = "10.42.0.0/16"
}

variable "availability_zone" {
  type    = string
  default = "ap-southeast-1a"
}

variable "admin_cidr" {
  description = "Trusted administrator CIDR for temporary bootstrap access; never use 0.0.0.0/0."
  type        = string
  nullable    = false
}
