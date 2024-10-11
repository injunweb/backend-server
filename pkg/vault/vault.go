package vault

import (
	"context"
	"fmt"

	"github.com/injunweb/backend-server/internal/config"

	"github.com/hashicorp/vault/api"
)

var (
	client *api.Client
	ctx    = context.Background()
)

func Init() error {
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = config.AppConfig.VaultAddr

	var err error
	client, err = api.NewClient(vaultConfig)
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %v", err)
	}

	client.SetToken(config.AppConfig.VaultToken)

	return nil
}

func GetSecret(path string) (map[string]interface{}, error) {
	secret, err := client.KVv1(config.AppConfig.VaultKV).Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read from Vault: %v", err)
	}

	return secret.Data, nil
}

func UpdateSecret(path string, data map[string]interface{}) error {
	err := client.KVv1(config.AppConfig.VaultKV).Put(ctx, path, data)
	if err != nil {
		return fmt.Errorf("failed to write to Vault: %v", err)
	}

	return nil
}

func InitSecret(path string, data map[string]interface{}) error {
	err := client.KVv1(config.AppConfig.VaultKV).Put(ctx, path, data)
	if err != nil {
		return fmt.Errorf("failed to initialize Vault secret: %v", err)
	}

	return nil
}

func DeleteSecret(path string) error {
	err := client.KVv1(config.AppConfig.VaultKV).Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %v", err)
	}

	return nil
}
