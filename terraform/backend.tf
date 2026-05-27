terraform {
  backend "s3" {
    bucket         = "jds-infra"
    key            = "ecs-fargate-lowcost/terraform.tfstate"
    region         = "us-west-1"
    encrypt        = true
    profile        = "jds"
  }
}
