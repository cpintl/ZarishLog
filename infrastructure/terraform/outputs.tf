output "database_endpoint" {
  description = "RDS database endpoint"
  value       = module.database.db_instance_address
}

output "redis_endpoint" {
  description = "Redis primary endpoint"
  value       = module.redis.primary_endpoint_address
}

output "ecr_api_repository_url" {
  description = "ECR API repository URL"
  value       = aws_ecr_repository.api.repository_url
}

output "ecr_web_repository_url" {
  description = "ECR Web repository URL"
  value       = aws_ecr_repository.web.repository_url
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.main.name
}

output "app_secrets_arn" {
  description = "Secrets Manager ARN"
  value       = aws_secretsmanager_secret.app.arn
}
