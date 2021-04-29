package template

import (
	"os"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
)

// Values are passed to templates while rendering.
type Values map[string]interface{}

// Merge merges other on top of v. Non-zero fields in other will override the
// same fields in v.
func (v Values) Merge(other Values) error {
	return mergo.Merge(&v, other, mergo.WithOverride)
}

// MergeValues merges values on top of each other from left to right. Returns a
// new Values map.
func MergeValues(values ...Values) (Values, error) {
	vals := Values{}

	for _, other := range values {
		err := vals.Merge(other)
		if err != nil {
			return nil, err
		}
	}

	return vals, nil
}

// LoadValues loads values from a file.
func LoadValues(path string) (Values, error) {
	var values Values

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buf, &values)
	if err != nil {
		return nil, err
	}

	return values, nil
}
