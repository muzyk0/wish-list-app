package models

// HandoffCode represents the result of generating a mobile handoff code.
type HandoffCode struct {
	Code      string
	ExpiresIn int // seconds until expiry
}
