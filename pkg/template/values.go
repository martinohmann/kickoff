package template

import "github.com/imdario/mergo"

// Values are passed to templates while rendering.
type Values map[string]interface{}

// Merge merges other on top of v. Non-zero fields in other will override the
// same fields in v.
func (v Values) Merge(other Values) error {
	return mergo.Merge(&v, other, mergo.WithOverride)
}

// MergeValues merges rhs on top of lhs without altering lhs. Returns a new
// Values map.
func MergeValues(lhs, rhs Values) (Values, error) {
	values := Values{}

	err := values.Merge(lhs)
	if err != nil {
		return nil, err
	}

	err = values.Merge(rhs)
	if err != nil {
		return nil, err
	}

	return values, nil
}
