package acc_shm

import "errors"

var (
	ErrReaderClosed        = errors.New("shared memory reader is closed")
	ErrInvalidPhysicsSize  = errors.New("invalid physics page size")
	ErrInvalidGraphicsSize = errors.New("invalid graphics page size")
	ErrInvalidStaticSize   = errors.New("invalid static page size")
	ErrUnsupportedPlatform = errors.New("shared memory reader is only supported on windows")
)
