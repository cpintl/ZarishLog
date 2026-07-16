terraform {
  required_version = ">= 1.8"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

locals {
  project     = "zarishlog"
  environment = var.environment
  name_prefix = "${local.project}-${local.environment}"
}

# VPC
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${local.name_prefix}-vpc"
  cidr = var.vpc_cidr

  azs             = var.availability_zones
  private_subnets = var.private_subnet_cidrs
  public_subnets  = var.public_subnet_cidrs

  enable_nat_gateway   = true
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = { Project = local.project, Environment = local.environment }
}

# RDS PostgreSQL
module "database" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 6.0"

  identifier = "${local.name_prefix}-db"

  engine               = "postgres"
  engine_version       = "18"
  family               = "postgres18"
  major_engine_version = "18"
  instance_class       = var.db_instance_class

  allocated_storage     = var.db_allocated_storage
  max_allocated_storage = var.db_max_allocated_storage
  storage_encrypted     = true

  db_name  = "zarishlog"
  username = "zarishlog"
  password = random_password.db_password.result
  port     = 5432

  vpc_security_group_ids = [module.vpc.default_security_group_id]
  db_subnet_group_name   = module.vpc.database_subnet_group

  backup_window      = "03:00-04:00"
  maintenance_window = "Mon:04:00-Mon:05:00"

  backup_retention_period = var.db_backup_retention_days
  deletion_protection     = var.environment == "production"

  tags = { Project = local.project, Environment = local.environment }
}

# ElastiCache Redis
module "redis" {
  source  = "terraform-aws-modules/elasticache/aws"
  version = "~> 1.0"

  cluster_id = "${local.name_prefix}-redis"

  engine         = "redis"
  engine_version = "8.0"
  node_type      = var.redis_node_type
  num_cache_nodes = 1

  subnet_group_name = module.vpc.database_subnet_group
  security_group_ids = [module.vpc.default_security_group_id]

  parameters = [
    { name = "notify-keyspace-events", value = "Ex" },
  ]

  tags = { Project = local.project, Environment = local.environment }
}

# ECS Fargate cluster
resource "aws_ecs_cluster" "main" {
  name = "${local.name_prefix}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = { Project = local.project, Environment = local.environment }
}

# ECR repositories
resource "aws_ecr_repository" "api" {
  name = "${local.name_prefix}-api"
  image_scanning_configuration {
    scan_on_push = true
  }
  tags = { Project = local.project, Environment = local.environment }
}

resource "aws_ecr_repository" "web" {
  name = "${local.name_prefix}-web"
  image_scanning_configuration {
    scan_on_push = true
  }
  tags = { Project = local.project, Environment = local.environment }
}

# Secrets
resource "random_password" "db_password" {
  length  = 32
  special = false
}

resource "aws_secretsmanager_secret" "app" {
  name = "${local.name_prefix}-secrets"
  tags = { Project = local.project, Environment = local.environment }
}

resource "aws_secretsmanager_secret_version" "app" {
  secret_id = aws_secretsmanager_secret.app.id
  secret_string = jsonencode({
    DATABASE_URL    = "postgresql://zarishlog:${random_password.db_password.result}@${module.database.db_instance_address}:5432/zarishlog"
    REDIS_URL       = "redis://${module.redis.primary_endpoint_address}:6379"
    JWT_SECRET      = random_password.jwt_secret.result
  })
}

resource "random_password" "jwt_secret" {
  length  = 64
  special = false
}
