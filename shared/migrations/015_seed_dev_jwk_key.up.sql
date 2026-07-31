-- up
UPDATE "user-service".jwk_key
SET active = FALSE
WHERE active = TRUE
  AND kid <> 'user-service-dev-rs256-20260507';

INSERT INTO "user-service".jwk_key (
    kid,
    alg,
    kty,
    public_pem,
    private_pem,
    active,
    created_at,
    rotated_at,
    expires_at
)
VALUES (
    'user-service-dev-rs256-20260507',
    'RS256',
    'RSA',
    $$-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArLwE53jTWYzl5Oqlvd27
vgotfZRC0qAESjNTwe8JuOOzT1QYM0WDpWR0RzYQE2UaSmZ5p/rmDLqGyOIAs5Ob
utmOCif1tCXwH0uPAScnfk8vs2f+VV9hclB1j1rJZMFXoRcm8kAiojU7TacYkgGt
4oPpXwpTM3M1NaN9Yuacz4bFczgjYjuYD5jiZNqSiFzBwWiFQq4WNQ5XVXWNm3Gm
66OPyY3MpUQkg0h+9UK6nGPCB6jil91368ggYvyXIbnSs4QFhBM6JSKgYX8eB67q
wzcXo0y15nzPb9WrjqHq9k7RqEEE5HXe/qQVXIgTQCjXUAjtbFJWgdAX8JP863SX
OQIDAQAB
-----END PUBLIC KEY-----$$,
    $$-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEArLwE53jTWYzl5Oqlvd27vgotfZRC0qAESjNTwe8JuOOzT1QY
M0WDpWR0RzYQE2UaSmZ5p/rmDLqGyOIAs5ObutmOCif1tCXwH0uPAScnfk8vs2f+
VV9hclB1j1rJZMFXoRcm8kAiojU7TacYkgGt4oPpXwpTM3M1NaN9Yuacz4bFczgj
YjuYD5jiZNqSiFzBwWiFQq4WNQ5XVXWNm3Gm66OPyY3MpUQkg0h+9UK6nGPCB6ji
l91368ggYvyXIbnSs4QFhBM6JSKgYX8eB67qwzcXo0y15nzPb9WrjqHq9k7RqEEE
5HXe/qQVXIgTQCjXUAjtbFJWgdAX8JP863SXOQIDAQABAoIBAAjLlhBXNaPEqdwL
Gp9dT/bwO7q+NtzUqwNAM86XJk6UwYeTh5vsuTRNtiH+HblvF3ScXStxeg9B3CUU
ZOa/6FkORM49lKQ0nlJpnYF4helHjO08qVWdgq+4axP+kmyf759TN3d0To8l2Lwu
evDMRxdWkiZ1tyDSh+4QQg0sIuXqKGavMOGKXSfyUmLAhCGDG1h7fl9pNCVRKFQw
w3FnLZigUdtmEvTcqpJmg/Cz4kvdn0TK0s1JoDqpHdbEoV9KPQgwslQSdc8F85R4
eBadsJwoun+xg2CjmPb9R3iv+bNhaNzhBDleUqLp73TWPNYT6tKXVI2AcBhtV94E
oYYZxwUCgYEA3n/sS8Udjz6T5FGd+JCqwzKxjN1HhDi/3zm3k277XrFCJymYnvml
2KpshHkZuZYCPdJs9tPDfVTFgF3EapGYvteaPHUiaST8y7IFh38jzvmcb6H8GPbw
6Im/4PAhXY3Ls1QsOk9F0fP71wt5YIxjjEVrnkCCL8pS2wx2CWpuaTsCgYEAxr3u
+RUU7bZJKstwPXJEwyJnlKxRgel86MbjNmH2Pmxn0NOEEoSZWmr+70SOCd3N+2Ro
N41yeBMcDmxKCb1/fWMHvcyFbHdONrIUW3Sw6XlBxOUBg4f7ehItibhl01h4XA/3
jgZNIDPPQ2W3pqI4wf0GKyCwpus9FL24kXtmmhsCgYEArc7qTKo3lB2DM/kZ2QFR
k9g24F4/LqeSIxOYNwCcNnVrwuH4ij9kcaN3z+g100a+i4KkghAchvxAqC0XcVQ5
KOONZaru7YnqPEjdjuIfm+Bbdszn/KxytoRcsp+CwO0ycezP++DPHtpkIbGh6Gzi
msHj9qRXznNTVDAgyOwuQd8CgYEAkKE2IKQb5+YJFxCXrM/UhKEr+gDxC/asBQZ/
4VqnBcSERG85JPTEWQ2WWu9r4ng8516pjQvtqr5VY5Wgx7fU6J3By3jj/AxSqfEs
aWXhPPcWSsBROrQh6TMDWr8LsyMl6/FeuUeSpwWtJqIGZUiWv21wKMCQbdixSb/L
amwAPdMCgYEAwV/UzgBF8dT+Yg8CTE+FqM03Bp03m+XKg3Vy7FpYJjzBkvu1UHeX
3cdNS1vFzOeB6w4hFpVxMVtzl+7j1cndcEzfkqc1qLirXk/EfRvrmhBqaqQtuvKA
g9mQHUOgxj8+aK99k2sNvChzLj3jt15KnnocFE5vJlxc8N9ylg84gL8=
-----END RSA PRIVATE KEY-----$$,
    TRUE,
    NOW(),
    NOW(),
    NOW() + INTERVAL '365 days'
)
ON CONFLICT (kid) DO UPDATE
SET alg = EXCLUDED.alg,
    kty = EXCLUDED.kty,
    public_pem = EXCLUDED.public_pem,
    private_pem = EXCLUDED.private_pem,
    active = TRUE,
    rotated_at = NOW(),
    expires_at = EXCLUDED.expires_at;
