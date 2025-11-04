package codec

// normalizedValue represents a value that has been normalized according to the
// TOON data model and is ready for emission by the encoder.
type normalizedValue interface{}

// numberValue captures a numeric literal that should be rendered verbatim.
type numberValue struct {
	literal string
}

// maxSafeInteger mirrors JavaScript's Number.MAX_SAFE_INTEGER, the threshold at
// which IEEE 754 double precision can no longer represent integers exactly.
const maxSafeInteger = 9007199254740991
