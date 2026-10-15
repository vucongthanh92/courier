# AWS development foundation

This Terraform root is pinned to Singapore (`ap-southeast-1`). It deliberately
starts with the network boundary only. It does not provision EC2, RDS, S3, ECR,
or IAM until their instance/database sizing, S3 retention, and cost estimate
are reviewed in the same pull request.

Run with a dedicated remote state backend before applying. Do not commit
`*.tfvars`, state files, or credentials.
