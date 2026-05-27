# Plan: Low-Cost ECS Fargate Deployment for poopjournal3 API

## Goals
- Run the Go API on AWS ECS using Fargate Spot for minimal cost
- Use S3 for all persistent storage (no EFS)
- Allow Route53 domain/cert to be provided as variables
- Support ARM/Graviton for compute savings
- Allow scaling to zero during idle periods
- Keep infra simple and maintainable

## Key AWS Resources
- **ECS Cluster** (Fargate capacity provider, Spot enabled)
- **ECS Task Definition** (ARM64, minimal CPU/mem, S3 env vars)
- **ECS Service** (min=0, max=1-2 tasks, autoscaling)
- **Application Load Balancer** (ALB, HTTPS, target group)
- **Route53 Record** (variable domain, points to ALB)
- **ACM Certificate** (variable ARN, for HTTPS)
- **IAM Roles/Policies** (for ECS task to access S3)
- **VPC, Subnets, Security Groups** (minimal, public ALB, private tasks)

## Cost Optimization
- Use Fargate Spot capacity provider for ECS service
- Set desired/minimum task count to 0 (scale to zero)
- Use ARM64 (Graviton) platform for task definition
- Smallest possible CPU/memory for Go API (e.g. 0.25 vCPU, 0.5GB RAM)
- No EFS, no RDS, no NAT Gateway (S3 only)

## Variables
- `domain_name` (Route53 record, e.g. api.example.com)
- `certificate_arn` (ACM cert for HTTPS)
- `s3_bucket` (existing or created outside this plan)

## Outputs
- ALB DNS name
- Service URL

## Terraform Structure
- `main.tf` — root module, ECS cluster, VPC, ALB, service
- `variables.tf` — input variables for domain, cert, S3
- `outputs.tf` — outputs for endpoints
- `ecs-task-def.json` — container/task definition (ARM, env vars)

## Next Steps
1. Write `variables.tf` for domain, cert, S3 bucket
2. Write `main.tf` for VPC, ECS, ALB, IAM, service
3. Write `outputs.tf` for ALB DNS, service URL
4. Write `ecs-task-def.json` for ARM64 Fargate Spot
5. Document how to deploy and update

---

This plan is designed for minimal cost and easy scaling. All stateful data is in S3. Route53 and ACM certs are expected to be managed outside this module and passed in as variables.
