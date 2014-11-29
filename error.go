package rgbe

import (
	"errors"
)

const (
	ReadError   = iota
	WriteError  = iota
	FormatError = iota
	MemoryError  = iota
)

func newError(code int, text string) error {
	switch code {
	case ReadError:
		return errors.New("RGBE read error: " + text)
	case WriteError:
		return errors.New("RGBE write error: " + text)
	case FormatError:
		return errors.New("RGBE bad file format: " + text)
	case MemoryError:
		fallthrough
	default:
		return errors.New("RGBE error: " + text)
	}
}