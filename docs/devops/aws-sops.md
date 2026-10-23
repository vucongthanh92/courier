# SOPS age secrets for AWS development

Courier AWS development uses SOPS with one age recipient. The public recipient
is in `.sops.yaml`; the private key is stored as an SSM SecureString named
`/courier/dev/argocd/sops-age-key` in `ap-southeast-2`.

Terraform intentionally does not manage the parameter value: Terraform state
must not become a second plaintext copy of the private key.

## Bootstrap the Argo CD key Secret

After the reviewed Terraform policy has been applied, run this through SSM on
the Application EC2. It retrieves the key on that host, creates or replaces
only the `argocd-sops-age` Secret in the `argocd` namespace, and removes its
temporary file before command completion.

```bash
export AWS_PROFILE=courier-terraform
export AWS_REGION=ap-southeast-2
export APPLICATION_INSTANCE_ID=i-06109722166dede79

aws ssm send-command \
  --region "$AWS_REGION" \
  --instance-ids "$APPLICATION_INSTANCE_ID" \
  --document-name AWS-RunShellScript \
  --comment 'Bootstrap Argo CD SOPS age key' \
  --parameters 'commands=["set -euo pipefail","key_file=$(mktemp /tmp/courier-sops-age.XXXXXX)","trap '\''rm -f \"$key_file\"'\'' EXIT","aws ssm get-parameter --region ap-southeast-2 --name /courier/dev/argocd/sops-age-key --with-decryption --query Parameter.Value --output text > \"$key_file\"","sudo /usr/local/bin/kubectl -n argocd create secret generic argocd-sops-age --from-file=keys.txt=\"$key_file\" --dry-run=client -o yaml | sudo /usr/local/bin/kubectl apply -f -","sudo /usr/local/bin/kubectl -n argocd get secret argocd-sops-age"]'
```

The command output must identify only the Secret metadata. Do not add a command
that prints the parameter value or Secret data.

## Rotation and recovery

Rotation requires a new age key pair, SSM parameter update, a re-encryption of
every `.enc.yaml` file with both recipients during migration, Argo CD key
Secret refresh, then removal of the old recipient. Test a decrypt/reconcile
before deleting the old key.

If the SSM parameter is lost, encrypted manifests cannot be decrypted. Restore
the parameter from a protected key backup before changing Git history.
