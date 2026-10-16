# Terraform state bootstrap

This root creates the encrypted, versioned, private S3 state bucket and its
on-demand DynamoDB lock table. Apply this root once, then configure the
`aws-dev` root to use the bucket and table as its remote backend.

It intentionally does not use a remote backend itself because it creates that
backend. Do not delete the bucket or lock table while any Terraform state uses
them.
