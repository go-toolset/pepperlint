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
