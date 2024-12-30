package filemanager

type DeleteType string

const (
	DeleteTypeNormal DeleteType = "Normal delete (faster, but recoverable)"
	DeleteTypeSecure DeleteType = "Secure delete (slower, but unrecoverable)"
)
