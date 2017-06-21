package comms

import (
	"errors"
)

var (
	EOF_AND_EXITED = errors.New("EOF and Exited")
)

func IsEOFAndExitedErr(err error) bool {
	if err == nil {
		return false
	}

	return err.Error() == EOF_AND_EXITED.Error()
}
