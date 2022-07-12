package xvalidator

// regexp pattern
const (
	nameRegex  = `[[:alpha:]][A-Za-z0-9_]*`
	constRegex = `[A-Z][A-Z0-9]*`
	fnRegex    = `(?m)([[:alpha:]][A-Za-z0-9_]*)\((.*?)\)`
)
