package redis

type Config struct {
	Addrs    []string
	Password string
	Username string
	DB       int
	PoolSize int
}
