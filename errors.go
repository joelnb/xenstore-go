package xenstore

import (
	"errors"
	"syscall"
)

var xenStoreErrors = map[string]syscall.Errno{
	"EINVAL":    syscall.EINVAL,
	"EACCES":    syscall.EACCES,
	"EEXIST":    syscall.EEXIST,
	"EISDIR":    syscall.EISDIR,
	"ENOENT":    syscall.ENOENT,
	"ENOMEM":    syscall.ENOMEM,
	"ENOSPC":    syscall.ENOSPC,
	"EIO":       syscall.EIO,
	"ENOTEMPTY": syscall.ENOTEMPTY,
	"ENOSYS":    syscall.ENOSYS,
	"EROFS":     syscall.EROFS,
	"EBUSY":     syscall.EBUSY,
	"EAGAIN":    syscall.EAGAIN,
	"EISCONN":   syscall.EISCONN,
}

// Error converts a string returned from XenStore to the syscall error
// it represents.
func Error(s string) error {
	if err, ok := xenStoreErrors[s]; ok {
		return err
	} else {
		return errors.New(s)
	}
}
