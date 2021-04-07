package kickoff

// Defaulter can set defaults for unset fields.
type Defaulter interface {
	// ApplyDefaults sets unset fields of the data structure to its default
	// values which might not necessarily be the zero value.
	ApplyDefaults()
}
