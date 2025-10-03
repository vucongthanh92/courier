package config

import (
	"os"
	"strings"

	"github.com/vucongthanh92/go-base-utils/vault"

	vaultgo "github.com/mittwald/vaultgo"
	"github.com/spf13/viper"
)

func LoadConfig(configPath string, config interface{}) {
	if configPath == "" {
		panic("Missing config path")
	}

	viper.SetConfigType(Yaml)
	viper.SetConfigFile(configPath)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	// update data from .vaultenv file
	if !viper.GetBool("development") {
		vaultEnvPath := os.Getenv("VAULT_ENV_PATH")
		if vaultEnvPath == "" {
			vaultEnvPath = "/app/.vaultenv"
		}
		println("Load config from external vault: " + vaultEnvPath + "")
		viper.SetConfigFile(vaultEnvPath)
		viper.SetConfigType(Env)
		if err := viper.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				if viper.GetString("vault.address") != "" {
					updateDataFromVault()
				}
			}

		}
	}
	if config != nil {
		if err := viper.Unmarshal(config); err != nil {
			panic(err)
		}
	}
}

func updateDataFromVault() {
	println("Fall back to update config from remote vault")
	address := viper.GetString("vault.address")
	path := viper.GetString("vault.path")
	token := viper.GetString("vault.token")
	role := viper.GetString("vault.role")
	mountPoint := viper.GetString("vault.mountPoint")

	if token == "" {
		token = os.Getenv("VAULT_TOKEN")
	}

	var vaultClient *vault.VaultClient
	if token != "" {
		vaultClient, _ = vault.NewVaultClient(address, vaultgo.WithAuthToken(token))
	} else {
		vaultClient, _ = vault.NewVaultClient(address, vaultgo.WithKubernetesAuth(role, vaultgo.WithMountPoint(mountPoint)))
	}

	secretData, err := vaultClient.GetSecretKeys(path)
	if err != nil {
		panic(err)
	}

	for key, s := range secretData {
		viper.Set(key, s)
	}
}
