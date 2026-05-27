output "alb_dns_name" {
  description = "The DNS name of the Application Load Balancer"
  value       = aws_lb.api_alb.dns_name
}

output "service_url" {
  description = "The public URL for the API service"
  value       = "https://${var.domain_name}"
}
