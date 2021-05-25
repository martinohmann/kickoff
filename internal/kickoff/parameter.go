package kickoff

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/martinohmann/kickoff/internal/template"
	"github.com/spf13/cast"
)

// ParameterType is the type for defining user-facing parameter type names.
type ParameterType string

const (
	StringParameterType     ParameterType = "string"
	NumberParameterType                   = "number"
	BoolParameterType                     = "bool"
	StringListParameterType               = "list<string>"
	NumberListParameterType               = "list<number>"
)

var validParameterTypes = []ParameterType{
	StringParameterType,
	NumberParameterType,
	BoolParameterType,
	StringListParameterType,
	NumberListParameterType,
}

// Parameters is a map of parameters keyed by name.
type Parameters map[string]*Parameter

// Validate implements the Validator interface.
func (p Parameters) Validate() error {
	for name, param := range p {
		if name == "" {
			return newParameterSpecError("name must not be empty")
		}

		if err := param.Validate(); err != nil {
			return fmt.Errorf("parameter %q: %w", name, err)
		}
	}

	return nil
}

// SetValues sets parameter values from a map keyed by parameter name. Returns
// an error if values contains an entry whose name does not exist in p or if
// setting the value of any parameter fails.
//
// See documentation of (*Parameter).SetValue(value).
func (p Parameters) SetValues(values map[string]interface{}) error {
	for name, val := range values {
		param, ok := p[name]
		if !ok {
			return fmt.Errorf("parameter %q not defined", name)
		}

		if err := param.SetValue(val); err != nil {
			return fmt.Errorf("parameter %q: %w", name, err)
		}
	}

	return nil
}

// Value returns a map of parameter names to parameter values. Returns an error
// if the value of any parameter cannot be looked up.
//
// See documentation of (*Parameter).Value().
func (p Parameters) Values() (template.Values, error) {
	values := make(template.Values)

	for name, param := range p {
		val, err := param.Value()
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", name, err)
		}

		values[name] = val
	}

	return values, nil
}

// ForEach calls fn for every key-value pair in p. Stops on the first error and
// returns it.
func (p Parameters) ForEach(fn func(string, *Parameter) error) error {
	if len(p) == 0 {
		return nil
	}

	keys := make([]string, 0, len(p))

	for k := range p {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		if err := fn(key, p[key]); err != nil {
			return err
		}
	}

	return nil
}

// Parameter is the schema for a user-defined skeleton parameter.
type Parameter struct {
	Type           ParameterType `json:"type"`
	AllowedValues  []interface{} `json:"allowedValues,omitempty"`
	AllowedPattern string        `json:"allowedPattern,omitempty"`
	Default        interface{}   `json:"default,omitempty"`
	Description    string        `json:"description,omitempty"`
	MinLength      *int          `json:"minLength,omitempty"`
	MaxLength      *int          `json:"maxLength,omitempty"`
	MinValue       *float64      `json:"minValue,omitempty"`
	MaxValue       *float64      `json:"maxValue,omitempty"`

	// Contains the parameter value if set.
	value interface{}
}

// Validate implements the Validator interface.
func (p *Parameter) Validate() error {
	if p == nil {
		return newParameterSpecError("must not be empty")
	}

	if err := validateParameterType(p.Type); err != nil {
		return err
	}

	switch p.Type {
	case NumberParameterType, NumberListParameterType:
		if p.MinValue != nil && p.MaxValue != nil && *p.MinValue > *p.MaxValue {
			return newParameterSpecError("minValue must be less or equal to maxValue")
		}
	case StringParameterType, StringListParameterType:
		if p.MinLength != nil && p.MaxLength != nil && *p.MinLength > *p.MaxLength {
			return newParameterSpecError("minLength must be less or equal to maxLength")
		}
	}

	if p.Default != nil {
		_, err := castAndValidateParameterValue(p, p.Default)
		if err != nil {
			return newParameterSpecError("invalid default value: %w", err)
		}
	}

	return nil
}

// SetValue sets the value of p. Returns an error if value is not convertible
// into the parameter's underlying type or if it violates any of the
// user-defined validation constraints.
func (p *Parameter) SetValue(value interface{}) error {
	v, err := castAndValidateParameterValue(p, value)
	if err != nil {
		return err
	}

	p.value = v
	return nil
}

// Value returns the parameter value. Returns the default value if no value was
// set via SetValue or an error if p does not have a default value.
func (p *Parameter) Value() (interface{}, error) {
	if p.value != nil {
		return p.value, nil
	}

	if p.Default != nil {
		return castAndValidateParameterValue(p, p.Default)
	}

	return nil, errors.New("value required")
}

func validateParameterType(typ ParameterType) error {
	for _, validType := range validParameterTypes {
		if typ == validType {
			return nil
		}
	}

	allowedValues := make([]string, len(validParameterTypes))
	for i, typ := range validParameterTypes {
		allowedValues[i] = string(typ)
	}

	return newParameterSpecError("invalid parameter type %q, allowed values: %s",
		typ, strings.Join(allowedValues, ", "))
}

func validateParameterValue(p *Parameter, value interface{}) error {
	switch tv := value.(type) {
	case bool:
		// no validation needed
	case float64:
		return validateNumberParameterValue(p, tv)
	case string:
		return validateStringParameterValue(p, tv)
	case []float64:
		for _, val := range tv {
			if err := validateParameterValue(p, val); err != nil {
				return err
			}
		}
	case []string:
		for _, val := range tv {
			if err := validateParameterValue(p, val); err != nil {
				return err
			}
		}
	default:
		panic(fmt.Errorf("unexpected parameter value: %#v", value))
	}

	return nil
}

func validateNumberParameterValue(p *Parameter, val float64) error {
	if p.MinValue != nil && val < *p.MinValue {
		return newParameterValueError("value must be >= %f, got: %f", *p.MinValue, val)
	}

	if p.MaxValue != nil && val > *p.MaxValue {
		return newParameterValueError("value must be <= %f, got: %f", *p.MaxValue, val)
	}

	if len(p.AllowedValues) > 0 {
		allowedValues, err := castToFloat64SliceE(p.AllowedValues)
		if err != nil {
			return err
		}

		if !containsFloat64(allowedValues, val) {
			return newParameterValueError("must be one of %s, got: %f",
				strings.Join(cast.ToStringSlice(p.AllowedValues), ", "), val)
		}
	}

	return nil
}

func validateStringParameterValue(p *Parameter, val string) error {
	if p.MinLength != nil && len(val) < *p.MinLength {
		return newParameterValueError("length must be >= %d, got: %s", *p.MinLength, val)
	}

	if p.MaxLength != nil && len(val) > *p.MaxLength {
		return newParameterValueError("length must be <= %d, got: %s", *p.MaxLength, val)
	}

	if len(p.AllowedValues) > 0 {
		allowedValues, err := cast.ToStringSliceE(p.AllowedValues)
		if err != nil {
			return err
		}

		if !containsString(allowedValues, val) {
			return newParameterValueError("must be one of %s, got: %s", strings.Join(allowedValues, ", "), val)
		}
	}

	if p.AllowedPattern != "" {
		pattern, err := regexp.Compile(p.AllowedPattern)
		if err != nil {
			return err
		}

		if !pattern.MatchString(val) {
			return newParameterValueError("must match pattern %s, got: %s", pattern, val)
		}
	}

	return nil
}

func castAndValidateParameterValue(p *Parameter, value interface{}) (interface{}, error) {
	v, err := castParameterValue(p.Type, value)
	if err != nil {
		return nil, err
	}

	if err := validateParameterValue(p, v); err != nil {
		return nil, err
	}

	return v, nil
}

func castParameterValue(typ ParameterType, value interface{}) (interface{}, error) {
	switch typ {
	case BoolParameterType:
		return cast.ToBoolE(value)
	case NumberParameterType:
		return cast.ToFloat64E(value)
	case NumberListParameterType:
		return castToFloat64SliceE(value)
	case StringParameterType:
		return cast.ToStringE(value)
	case StringListParameterType:
		return cast.ToStringSliceE(value)
	default:
		panic(fmt.Errorf("unexpected parameter type: %s", typ))
	}
}

func castToFloat64SliceE(i interface{}) ([]float64, error) {
	if i == nil {
		return []float64{}, fmt.Errorf("unable to cast %#v of type %T to []float64", i, i)
	}

	switch v := i.(type) {
	case []float64:
		return v, nil
	}

	kind := reflect.TypeOf(i).Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(i)
		a := make([]float64, s.Len())
		for j := 0; j < s.Len(); j++ {
			val, err := cast.ToFloat64E(s.Index(j).Interface())
			if err != nil {
				return []float64{}, fmt.Errorf("unable to cast %#v of type %T to []float64", i, i)
			}
			a[j] = val
		}
		return a, nil
	default:
		return []float64{}, fmt.Errorf("unable to cast %#v of type %T to []float64", i, i)
	}
}

func containsFloat64(haystack []float64, needle float64) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}

func containsString(haystack []string, needle string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}
