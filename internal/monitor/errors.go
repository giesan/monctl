package monitor

import "errors"

var ErrUnsupported = errors.New("brightness control is not supported")

func IsUnsupported(err error) bool {
	return errors.Is(err, ErrUnsupported)
}
