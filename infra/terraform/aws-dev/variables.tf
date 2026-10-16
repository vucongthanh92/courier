variable "aws_region" {
  type    = string
  default = "ap-southeast-2"
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
  default = "ap-southeast-2a"
}

variable "admin_cidr" {
  description = "Reserved for a future explicit administrative ingress rule. SSM is the only administrative access in this baseline."
  type        = string
  default     = null
}

variable "budget_alert_email" {
  description = "Email address that receives the AWS Budget and disk-usage alert subscriptions."
  type        = string
  nullable    = false
}

variable "application_instance_type" {
  description = "ARM64 EC2 instance type for K3s, Argo CD, apps, and compact monitoring."
  type        = string
  default     = "t4g.large"
}

variable "data_instance_type" {
  description = "ARM64 EC2 instance type for PostgreSQL, Redis, Kafka, and Qdrant."
  type        = string
  default     = "t4g.medium"
}

variable "application_root_volume_size" {
  type    = number
  default = 25
}

variable "data_root_volume_size" {
  type    = number
  default = 25
}

variable "data_volume_size" {
  type    = number
  default = 60
}

variable "k3s_version" {
  description = "Pinned K3s release installed on the application node."
  type        = string
  default     = "v1.37.1+k3s1"
}

variable "github_owner" {
  description = "GitHub owner whose main-branch workflows may assume the ECR publishing role."
  type        = string
  default     = "vucongthanh92"
}

variable "ecr_repositories" {
  description = "Courier service repositories published to private ECR."
  type        = set(string)
  default = [
    "user-service",
    "chat-service",
    "payment-gateway",
    "agent-gateway",
    "conversa-app",
    "flikk-app",
  ]
}

variable "github_repositories" {
  description = "Exact GitHub repositories allowed to publish Courier images from main."
  type        = set(string)
  default = [
    "courier-user-svc",
    "courier-chat-svc",
    "courier-payment-gateway",
    "courier-agent-gateway",
    "courier-conversa-app",
    "courier-flikk-app",
  ]
}
