# AWS development foundation

This Terraform root is pinned to Sydney (`ap-southeast-2`). The agreed
baseline is an application `t4g.large`, a data `t4g.medium`, 110 GiB gp3 across
three volumes, PostgreSQL on the data node, and separate files/backups buckets.
Instances are expected to run about 120 hours/month each. It deliberately
starts with the network boundary only; the remaining resources will be added
only with their cost and security controls.

Run with a dedicated remote state backend before applying. Do not commit
`*.tfvars`, state files, or credentials.
