package comms

import (
	"errors"
)

var (
	EOF_AND_EXITED_SUCCESSFULLY = errors.New("EOF and Exited successfully")
)

func IsEOFAndExitedSuccessfullyErr(err error) bool {
	if err == nil {
		return false
	}

	return err.Error() == EOF_AND_EXITED_SUCCESSFULLY.Error()
}
