package nkdb

const (
	ErrInternal = iota + 1
	ErrInvalid
	ErrNotExist
)

func Is(aError error, aType int) bool {
	if aError == nil {
		return false
	}

	tError, tOK := aError.(*Error)
	if !tOK {
		return false
	}

	return tError.Type == aType
}

func IsInternal(aError error) bool {
	return Is(aError, ErrInternal)
}

func IsInvalid(aError error) bool {
	return Is(aError, ErrInvalid)
}

func IsNotExist(aError error) bool {
	return Is(aError, ErrNotExist)
}

type Error struct {
	Type int
	Op   string
	Key  string
	Err  error
}

func (aError *Error) Error() string {
	if aError.Key == "" {
		return aError.Op + ": " + aError.Err.Error()
	}
	return aError.Op + " " + aError.Key + ": " + aError.Err.Error()
}

func (aError *Error) Unwrap() error {
	return aError.Err
}
