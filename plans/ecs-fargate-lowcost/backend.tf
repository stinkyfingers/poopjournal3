terraform {
  backend "s3" {
    bucket         = "<REPLACE_WITH_STATE_BUCKET>"
    key            = "ecs-fargate-lowcost/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "<OPTIONAL_LOCK_TABLE>" # Optional for state locking
    encrypt        = true
  }
}

# Instructions:
# 1. Replace <REPLACE_WITH_STATE_BUCKET> with your S3 bucket for state.
# 2. (Optional) Replace <OPTIONAL_LOCK_TABLE> with a DynamoDB table for state locking.
# 3. Run `terraform init` in this directory to initialize remote state.
