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
variable "certificate_arn" {
  description = "The ACM certificate ARN for HTTPS"
  type        = string
  default     = "arn:aws:acm:us-east-1:671958020402:certificate/3e254b04-1362-4b2e-a5a2-72f8e8a8b85e"
}
variable "s3_bucket" {
  description = "The S3 bucket name for API data storage"
  type        = string
  default     = "poopjournal-data-prod"
}
