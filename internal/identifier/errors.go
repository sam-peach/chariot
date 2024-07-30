package identifier

type errReason string

const (
	BadLength errReason = "Incorrect length."
	BadFormat errReason = "Incorrect format."
)

type ValidationError struct {
	reason errReason
}

func (e ValidationError) Error() string {
	return "Invalid identifier: " + string(e.reason)
}
