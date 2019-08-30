package consts

import "errors"

const (
	CodeOK         = 0
	CodeEtcdErr    = 1
	CodeMarshalErr = 2
	CodeNotFound   = 3
	CodeNotExisted = 7
	LevelSSD       = "SSD"
	LevelNVM       = "NVM"
	LevelHDD       = "HDD"
	KindCapability = "capability"
	KindAllocation = "allocation"
	KindRemaining  = "remaining"
	OpAdd          = int32(1)
	OpDel          = int32(0)
)

var (
	ErrNotExist = errors.New("key not exist")
	ErrNotFound = errors.New("pv not found")
	ErrNotBound = errors.New("pv not bounded")
	ErrRetry    = errors.New("retry")
)
