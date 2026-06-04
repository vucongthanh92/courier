package vault

import (
	"fmt"

	vaultgo "github.com/mittwald/vaultgo"
)

type VaultClient struct {
	*vaultgo.Client
}

func NewVaultClient(address string, opt vaultgo.ClientOpts) (*VaultClient, error) {
	client, err := vaultgo.NewClient(address, vaultgo.WithCaPath(""), opt)

	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	return &VaultClient{client}, nil
}

func (c *VaultClient) GetSecretKeys(path string) (map[string]interface{}, error) {
	secret, err := c.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data type assertion failed: %T %#v", secret.Data["data"], secret.Data["data"])
	}

	return data, nil
}

func (c *VaultClient) GetSecretKey(path, key string) (string, error) {
	data, _ := c.GetSecretKeys(path)

	value, ok := data[key].(string)
	if !ok {
		return "", fmt.Errorf("value type assertion failed: %T %#v", data[key], data[key])
	}

	return value, nil
}
