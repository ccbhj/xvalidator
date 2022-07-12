package internal

// regexp pattern
const (
	NameRegex  = `[[:alpha:]][A-Za-z0-9_]*`
	ConstRegex = `[A-Z][A-Z0-9]*`
	FnRegex    = `(?m)([[:alpha:]][A-Za-z0-9_]*)\((.*?)\)`
)
