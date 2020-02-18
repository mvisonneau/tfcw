package schemas

type S5 struct {
	CipherEngineType  *S5CipherEngineType  `hcl:"engine"`
	CipherEngineAES   *S5CipherEngineAES   `hcl:"aes,block"`
	CipherEngineAWS   *S5CipherEngineAWS   `hcl:"aws,block"`
	CipherEngineGCP   *S5CipherEngineGCP   `hcl:"gcp,block"`
	CipherEnginePGP   *S5CipherEnginePGP   `hcl:"pgp,block"`
	CipherEngineVault *S5CipherEngineVault `hcl:"vault,block"`
	Value             *string              `hcl:"value"`
}

type S5CipherEngineType string

const (
	S5CipherEngineTypeAES   S5CipherEngineType = "aes"
	S5CipherEngineTypeAWS   S5CipherEngineType = "aws"
	S5CipherEngineTypeGCP   S5CipherEngineType = "gcp"
	S5CipherEngineTypePGP   S5CipherEngineType = "pgp"
	S5CipherEngineTypeVault S5CipherEngineType = "vault"
)

type S5CipherEngineAES struct {
	Key *string `hcl:"key"`
}

type S5CipherEngineAWS struct {
	KmsKeyArn *string `hcl:"kms-key-arn"`
}

type S5CipherEngineGCP struct {
	KmsKeyName *string `hcl:"kms-key-name"`
}

type S5CipherEnginePGP struct {
	PublicKeyPath  *string `hcl:"public-key-path"`
	PrivateKeyPath *string `hcl:"private-key-path"`
}

type S5CipherEngineVault struct {
	TransitKey *string `hcl:"transit-key"`
}
