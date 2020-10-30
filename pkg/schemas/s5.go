package schemas

// S5 is a provider type
type S5 struct {
	CipherEngineType  *S5CipherEngineType  `hcl:"engine"`
	CipherEngineAES   *S5CipherEngineAES   `hcl:"aes,block"`
	CipherEngineAWS   *S5CipherEngineAWS   `hcl:"aws,block"`
	CipherEngineGCP   *S5CipherEngineGCP   `hcl:"gcp,block"`
	CipherEnginePGP   *S5CipherEnginePGP   `hcl:"pgp,block"`
	CipherEngineVault *S5CipherEngineVault `hcl:"vault,block"`
	Value             *string              `hcl:"value"`
}

// S5CipherEngineType represents a S5 cipher engine type
type S5CipherEngineType string

const (
	// S5CipherEngineTypeAES refers to an 'aes' s5 cipher engine type
	S5CipherEngineTypeAES S5CipherEngineType = "aes"

	// S5CipherEngineTypeAWS refers to an 'aws' s5 cipher engine type
	S5CipherEngineTypeAWS S5CipherEngineType = "aws"

	// S5CipherEngineTypeGCP refers to a 'gcp' s5 cipher engine type
	S5CipherEngineTypeGCP S5CipherEngineType = "gcp"

	// S5CipherEngineTypePGP refers to a 'pgp' s5 cipher engine type
	S5CipherEngineTypePGP S5CipherEngineType = "pgp"

	// S5CipherEngineTypeVault refers to a 'vault' s5 cipher engine type
	S5CipherEngineTypeVault S5CipherEngineType = "vault"
)

// S5CipherEngineAES handles necessary configuration for an 'aes' s5 cipher engine
type S5CipherEngineAES struct {
	Key *string `hcl:"key"`
}

// S5CipherEngineAWS handles necessary configuration for an 'aws' s5 cipher engine
type S5CipherEngineAWS struct {
	KmsKeyArn *string `hcl:"kms-key-arn"`
}

// S5CipherEngineGCP handles necessary configuration for a 'gcp' s5 cipher engine
type S5CipherEngineGCP struct {
	KmsKeyName *string `hcl:"kms-key-name"`
}

// S5CipherEnginePGP handles necessary configuration for a 'pgp' s5 cipher engine
type S5CipherEnginePGP struct {
	PublicKeyPath  *string `hcl:"public-key-path"`
	PrivateKeyPath *string `hcl:"private-key-path"`
}

// S5CipherEngineVault handles necessary configuration for a 'vault' s5 cipher engine
type S5CipherEngineVault struct {
	TransitKey *string `hcl:"transit-key"`
}
