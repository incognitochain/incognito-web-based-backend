package submitproof

const (
	ShieldStatusUnknown = iota
	ShieldStatusSubmitting
	ShieldStatusSubmitFailed
	ShieldStatusSubmitted
	ShieldStatusPending
	ShieldStatusRejected
	ShieldStatusAccepted
)

const (
	ShieldErrorPrefix = "error-shield-"
	ShieldStatusrefix = "status-shield-"
)
