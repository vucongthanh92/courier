# Courier System Documentation

# Run Database Migration

`migrate \
  -path migrations \
  -database "postgresql://dev:dev@127.0.0.1:5432/dev_db?sslmode=disable" \
  up
`