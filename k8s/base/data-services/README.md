# Data EC2 service contracts

These `ExternalName` Services make the private Data EC2 available to pods in
`courier-dev` under stable names:

| Service | Destination | Consumers |
| --- | --- | --- |
| `postgresql:5432` | Data EC2 PostgreSQL | user, chat, payment |
| `redis:6379` | Data EC2 Redis | user, chat, payment |
| `kafka:9092` | Data EC2 Kafka broker | user, chat, payment, agent |
| `qdrant:6333` | Data EC2 Qdrant HTTP API | agent |

The K3s CoreDNS custom configuration owns the mapping from
`data.courier.internal` to Data EC2's Terraform-assigned private address
`10.42.0.114`. A stop/start retains it; replacing the instance does not. Update
the CoreDNS mapping in the same reviewed change if that instance is replaced.

The Data security group accepts these four ports only from the Application EC2
security group. These Services do not expose a public load balancer or bypass
the AWS network boundary. Argo CD deliberately excludes `EndpointSlice`; this
design keeps the entire service discovery contract under GitOps ownership.
