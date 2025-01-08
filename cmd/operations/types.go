package operations

type OperationType string

const (
	OperationEncrypt OperationType = "encryption"
	OperationDecrypt OperationType = "decryption"
)

type OperationConfig struct {
	InputPath  string
	OutputPath string
	Password   string
	Operation  OperationType
}
