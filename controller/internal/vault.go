package internal

import (
	"context"
	"log"

	vault "github.com/hashicorp/vault/api"
)

func GetAWSCreds() (string, string, string) {
	config := vault.DefaultConfig()
	config.Address = "http://vault:8200"

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("Unable to initialize a Vault client: %v", err)
	}

	client.SetToken("hvs.LmVCy2Qd3wUxMsD6IBzByCFQ")

	secret, err := client.KVv2("kv").Get(context.Background(), "aws/credentials")
	if err != nil {
		log.Fatalf(
			"Unable to read the super secret password from the vault: %v",
			err,
		)
	}

	accessKey, ok := secret.Data["AWS_ACCESS_KEY_ID"].(string)
	if !ok {
		log.Fatalf(
			"value type assertion failed: %T %#v",
			secret.Data["AWS_ACCESS_KEY_ID"],
			secret.Data["AWS_ACCESS_KEY_ID"],
		)
	}

	secretKey, ok := secret.Data["AWS_SECRET_ACCESS_KEY"].(string)
	if !ok {
		log.Fatalf(
			"value type assertion failed: %T %#v",
			secret.Data["AWS_SECRET_ACCESS_KEY"],
			secret.Data["AWS_SECRET_ACCESS_KEY"],
		)
	}

	region, ok := secret.Data["AWS_REGION"].(string)
	if !ok {
		log.Fatalf(
			"value type assertion failed: %T %#v",
			secret.Data["AWS_REGION"],
			secret.Data["AWS_REGION"],
		)
	}

	return accessKey, secretKey, region
}
