package template

import "github.com/imdario/mergo"

// Values are passed to templates while rendering.
type Values map[string]interface{}

// Merge merges other on top of v. Non-zero fields in other will override the
// same fields in v.
func (v Values) Merge(other Values) error {
	return mergo.Merge(&v, other, mergo.WithOverride)
}
