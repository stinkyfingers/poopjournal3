variable "profile" {
  description = "The AWS CLI profile to use for authentication (optional)"
  type        = string
  default     = "jds"
}
variable "aws_region" {
  description = "The AWS region to deploy resources into"
  type        = string
  default     = "us-west-1"
}
variable "domain_name" {
  description = "The Route53 domain name for the API (e.g. api.example.com)"
  type        = string
  default     = "api.poopjournal.com"
}
variable "hosted_zone_name" {
  description = "The Route53 hosted zone name (e.g. example.com)"
  type        = string
  default     = "poopjournal.com"
}
variable "certificate_arn" {
  description = "The ACM certificate ARN for HTTPS"
  type        = string
  default     = "arn:aws:acm:us-west-1:671958020402:certificate/1571bdf4-b591-4f2a-bddf-863ec6d0da87"
}
variable "s3_bucket" {
  description = "The S3 bucket name for API data storage"
  type        = string
  default     = "poopjournal-data-prod"
}
