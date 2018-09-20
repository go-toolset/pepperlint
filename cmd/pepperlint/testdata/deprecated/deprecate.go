package deprecated

// Deprecated structure
//
// Deprecated: use Foo instead
type Deprecated struct {
	// DeprecatedField is deprecated
	//
	// Deprecated: Use Field instead
	DeprecatedField int32

	Field int64
}

// Foo structure is a non-deprecated structure
type Foo struct {
	Field float64
}

// DeprecatedOp operation
//
// Deprecated: Use Op instead
func (d Deprecated) DeprecatedOp() int {
	return 1
}

// DeprecatedPtrOp operation
//
// Deprecated: Use Op instead
func (d *Deprecated) DeprecatedPtrOp() int {
	return 0
}

// Op valid operation
func (d Deprecated) Op() {
}

// DeprecatedFunction operation
//
// Deprecated: Use Function instead
func DeprecatedFunction() {
}

// Function is a valid function
func Function() {
}
