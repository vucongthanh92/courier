# AWS development foundation

This Terraform root is pinned to Sydney (`ap-southeast-2`) and uses the
bootstrap-created remote state bucket plus DynamoDB lock table. The agreed
baseline is an application `t4g.large`, a data `t4g.medium`, 110 GiB gp3 across
three volumes, PostgreSQL on the data node, and separate files/backups buckets.
Instances are expected to run about 120 hours/month each.

It provisions the AWS foundation only: VPC/security groups, ARM64 EC2 nodes,
EBS, private ECR repositories, files/backups buckets, SSM access, GitHub OIDC
image-publisher access, and cost/disk alerts. It does not store application
credentials, deploy stateful services, publish images, or enable public HTTP.

Both instances use the public subnet because this cost-constrained design has
no NAT gateway. The data security group has **no inbound Internet rule**; it
accepts PostgreSQL, Redis, Kafka, and Qdrant only from the application security
group. Its public IP exists solely for outbound package/SSM access and may
change after a stop/start cycle.

Before the first plan, migrate the prior local state to the remote backend:

```bash
AWS_PROFILE=courier-terraform terraform -chdir=infra/terraform/aws-dev init -migrate-state
```

Create a local ignored `dev.tfvars` with your alert email:

```hcl
budget_alert_email = "you@example.com"
```

Then run `terraform plan -var-file=dev.tfvars -out=aws-dev.tfplan` and review
the resource count and all public boundaries before applying. Do not commit
`*.tfvars`, state files, plans, or credentials.
