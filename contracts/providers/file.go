package providerContract

import (
	"io"
)

type File interface {
	Test() error
	SaveFile(io.Reader, string) error
	ReadFile(string) (io.ReadCloser, error)
	DeleteFile(string) error
}
