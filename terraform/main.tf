terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
  }
  required_version = ">= 1.3"
}

provider "aws" {
  region  = var.aws_region
  profile = var.profile
}

# VPC (small, 2 AZs, public+private)
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.1.0"
  name    = "poopjournal-vpc"
  cidr    = "10.0.0.0/16"
  azs     = slice(data.aws_availability_zones.available.names, 0, 2)
  public_subnets  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnets = ["10.0.101.0/24", "10.0.102.0/24"]
  enable_nat_gateway = false
  single_nat_gateway = false
  enable_dns_hostnames = true
  # Removed unsupported 'vpc = true' argument (not valid in AWS provider v3+)
}

data "aws_availability_zones" "available" {}

# ECS Cluster with Fargate Spot
resource "aws_ecs_cluster" "api" {
  name = "poopjournal-api"
  setting {
    name  = "containerInsights"
    value = "disabled"
  }
}

resource "aws_ecs_cluster_capacity_providers" "api" {
  cluster_name = aws_ecs_cluster.api.name
  capacity_providers = ["FARGATE", "FARGATE_SPOT"]
  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
    base              = 0
  }
}

# IAM Role for ECS Task
resource "aws_iam_role" "ecs_task_execution" {
  name = "poopjournal-ecs-task-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_assume_role.json
}

data "aws_iam_policy_document" "ecs_task_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_policy" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy" "ecs_task_s3" {
  name = "poopjournal-ecs-task-s3"
  role = aws_iam_role.ecs_task_execution.id
  policy = data.aws_iam_policy_document.ecs_task_s3.json
}

data "aws_iam_policy_document" "ecs_task_s3" {
  statement {
    actions = ["s3:GetObject", "s3:PutObject", "s3:ListBucket"]
    resources = [
      "arn:aws:s3:::${var.s3_bucket}",
      "arn:aws:s3:::${var.s3_bucket}/*"
    ]
  }
}

# ALB
resource "aws_lb" "api_alb" {
  name               = "poopjournal-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = module.vpc.public_subnets
}

resource "aws_lb_target_group" "api_tg" {
  name     = "poopjournal-tg"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = module.vpc.vpc_id
  target_type = "ip"
  health_check {
    path                = "/status"
    matcher             = "200-399"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.api_alb.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = var.certificate_arn
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api_tg.arn
  }
}

resource "aws_cloudwatch_log_group" "ecs_api" {
  name              = "/ecs/poopjournal-api"
  retention_in_days = 14
}

# Security Groups
resource "aws_security_group" "alb" {
  name        = "poopjournal-alb-sg"
  description = "Allow HTTPS inbound"
  vpc_id      = module.vpc.vpc_id
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "ecs_tasks" {
  name        = "poopjournal-ecs-tasks-sg"
  description = "Allow ECS tasks to receive traffic from ALB"
  vpc_id      = module.vpc.vpc_id
  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# ECS Task Definition (from file)
resource "aws_ecs_task_definition" "api" {
  family                   = "poopjournal-api"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task_execution.arn
  container_definitions    = file("${path.module}/ecs-task-def.json")
  runtime_platform {
    cpu_architecture        = "ARM64"
    operating_system_family = "LINUX"
  }
}

# ECS Service (Fargate Spot, min=0, max=2)
resource "aws_ecs_service" "api" {
  name            = "poopjournal-api"
  cluster         = aws_ecs_cluster.api.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = 1
  network_configuration {
    subnets          = module.vpc.public_subnets
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = true
  }
  load_balancer {
    target_group_arn = aws_lb_target_group.api_tg.arn
    container_name   = "poopjournal-api"
    container_port   = 8080
  }
  capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
    base              = 0
  }
  depends_on = [aws_cloudwatch_log_group.ecs_api]
}

resource "aws_route53_record" "api" {
  zone_id = data.aws_route53_zone.selected.zone_id
  name    = var.domain_name
  type    = "A"
  alias {
    name                   = aws_lb.api_alb.dns_name
    zone_id                = aws_lb.api_alb.zone_id
    evaluate_target_health = true
  }
}

data "aws_route53_zone" "selected" {
  name         = var.hosted_zone_name
  private_zone = false
}
