# Encrypted Kubernetes secrets

Only SOPS-encrypted files ending in `.enc.yaml` belong in this directory.
The age private key is not in Git; it is stored as the AWS Systems Manager
SecureString `/courier/dev/argocd/sops-age-key`.

The AWS development public recipient is:

```text
age1gp9hfgqxdz2g3r7w0l3pufy7mt62gtp5g22jcyj7nuc0euwvfyhsfgpduu
```

Create plaintext secret material outside this repository, encrypt it before it
is copied here, and verify that no plaintext is staged:

```bash
sops --encrypt --in-place path/to/secret.enc.yaml
git diff --cached -- k8s/secrets
```

Do not use `sops --decrypt --in-place` in this repository. Decrypt to a
temporary file with restrictive permissions when local inspection is required.

Argo CD receives the private key only through its bootstrap Kubernetes Secret.
The repository and pull requests must never contain the private age key,
decrypted database URLs, API keys, OAuth credentials, or payment credentials.
