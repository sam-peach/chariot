package utils

type ApiError struct {
	Err error
}

func (e ApiError) Error() string {
	return "API error: " + e.Err.Error()
}

type RequestError struct {
	Err error
}

func (e RequestError) Error() string {
	return "Bad request: " + e.Err.Error()
}

type ApiErrReason int

const (
	Internal ApiErrReason = iota
)

func (e ApiErrReason) Error() string {
	switch e {
	case Internal:
		return "internal error"
	default:
		return ""
	}
}
