package kickoff

import (
	"errors"
	"testing"

	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/require"
)

func TestParameters_Validate(t *testing.T) {
	testCases := []struct {
		name string
		p    Parameters
		err  error
	}{
		{
			name: "nil map is valid",
		},
		{
			name: "empty parameter name is not allowed",
			p:    Parameters{"": &Parameter{}},
			err:  newParameterSpecError("name must not be empty"),
		},
		{
			name: "empty parameter body is not allowed",
			p:    Parameters{"foo": nil},
			err:  errors.New(`parameter "foo": invalid parameter spec: must not be empty`),
		},
		{
			name: "invalid parameter type",
			p:    Parameters{"foo": &Parameter{Type: "bar"}},
			err:  errors.New(`parameter "foo": invalid parameter spec: invalid parameter type "bar", allowed values: string, number, bool, list<string>, list<number>`),
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := test.p.Validate()
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParameter_Validate(t *testing.T) {
	testCases := []struct {
		name string
		p    Parameter
		err  error
	}{
		{
			name: "invalid parameter type",
			p:    Parameter{Type: "bar"},
			err:  newParameterSpecError(`invalid parameter type "bar", allowed values: string, number, bool, list<string>, list<number>`),
		},
		{
			name: "simple parameter is valid",
			p:    Parameter{Type: StringParameterType},
		},
		{
			name: "list parameter is valid",
			p:    Parameter{Type: StringListParameterType},
		},
		{
			name: "validates number ranges",
			p:    Parameter{Type: NumberParameterType, MinValue: float64Ptr(10), MaxValue: float64Ptr(9)},
			err:  newParameterSpecError(`minValue must be less or equal to maxValue`),
		},
		{
			name: "validates string length constraints",
			p:    Parameter{Type: StringParameterType, MinLength: intPtr(10), MaxLength: intPtr(9)},
			err:  newParameterSpecError(`minLength must be less or equal to maxLength`),
		},
		{
			name: "validates default value",
			p:    Parameter{Type: StringParameterType, Default: "foo"},
		},
		{
			name: "default value must have compatible type",
			p:    Parameter{Type: BoolParameterType, Default: "foo"},
			err:  newParameterSpecError(`invalid default value: strconv.ParseBool: parsing "foo": invalid syntax`),
		},
		{
			name: "default value must satify constraints",
			p:    Parameter{Type: StringParameterType, Default: "foo", MinLength: intPtr(10)},
			err:  newParameterSpecError(`invalid default value: invalid parameter value: length must be >= 10, got: foo`),
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := test.p.Validate()
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParameter_SetValues(t *testing.T) {
	testCases := []struct {
		name     string
		p        Parameters
		given    map[string]interface{}
		expected template.Values
		err      error
	}{
		{
			name: "setting undefined parameter fails",
			p: Parameters{
				"foo": &Parameter{Type: StringParameterType},
			},
			given: map[string]interface{}{"bar": "baz"},
			err:   errors.New(`parameter "bar" not defined`),
		},
		{
			name: "setting defined param succeeds",
			p: Parameters{
				"foo": &Parameter{Type: StringParameterType},
			},
			given:    map[string]interface{}{"foo": "baz"},
			expected: map[string]interface{}{"foo": "baz"},
		},
		{
			name: "parameter validation fails",
			p: Parameters{
				"foo": &Parameter{Type: StringParameterType, MaxLength: intPtr(2)},
			},
			given: map[string]interface{}{"foo": "baz"},
			err:   errors.New(`parameter "foo": invalid parameter value: length must be <= 2, got: baz`),
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := test.p.SetValues(test.given)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)

				got, err := test.p.Values()
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestParameter_SetValue(t *testing.T) {
	testCases := []struct {
		name     string
		p        Parameter
		given    interface{}
		expected interface{}
		err      error
	}{
		{
			name:     "set exact type",
			p:        Parameter{Type: StringParameterType},
			given:    "foo",
			expected: "foo",
		},
		{
			name:     "casts simple types",
			p:        Parameter{Type: NumberParameterType},
			given:    "1",
			expected: float64(1),
		},
		{
			name:     "casts lists",
			p:        Parameter{Type: NumberListParameterType},
			given:    []interface{}{"1", 2, 3.0},
			expected: []float64{1, 2, 3},
		},
		{
			name:     "casts lists #2",
			p:        Parameter{Type: StringListParameterType},
			given:    []interface{}{"1", 2, 3.0},
			expected: []string{"1", "2", "3"},
		},
		{
			name:  "value must satify length constraint",
			p:     Parameter{Type: StringParameterType, MaxLength: intPtr(2)},
			given: "foo",
			err:   newParameterValueError(`length must be <= 2, got: foo`),
		},
		{
			name:  "numeric value must satify range constraint",
			p:     Parameter{Type: NumberParameterType, MaxValue: float64Ptr(2.5)},
			given: 3,
			err:   newParameterValueError(`value must be <= 2.500000, got: 3.000000`),
		},
		{
			name:  "numeric value must satify range constraint #2",
			p:     Parameter{Type: NumberParameterType, MinValue: float64Ptr(42)},
			given: 3,
			err:   newParameterValueError(`value must be >= 42.000000, got: 3.000000`),
		},
		{
			name:  "value must match pattern",
			p:     Parameter{Type: StringParameterType, AllowedPattern: "^foo.*"},
			given: "bar",
			err:   newParameterValueError(`must match pattern ^foo.*, got: bar`),
		},
		{
			name:     "value must matches pattern",
			p:        Parameter{Type: StringParameterType, AllowedPattern: "^foo.*"},
			given:    "foobar",
			expected: "foobar",
		},
		{
			name:  "validates each list item",
			p:     Parameter{Type: StringListParameterType, AllowedValues: []interface{}{"foo", "bar"}},
			given: []string{"foo", "bar", "baz"},
			err:   newParameterValueError(`must be one of foo, bar, got: baz`),
		},
		{
			name:  "validates each numeric list item",
			p:     Parameter{Type: NumberListParameterType, AllowedValues: []interface{}{1, 2.5, 3}},
			given: []interface{}{"2.5", "3", 42},
			err:   newParameterValueError(`must be one of 1, 2.5, 3, got: 42.000000`),
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := test.p.SetValue(test.given)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)

				got, err := test.p.Value()
				require.NoError(t, err)
				require.Equal(t, test.expected, got)
			}
		})
	}
}

func TestParameter_Value(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		p := Parameter{Type: StringParameterType, Default: "foo"}

		val, err := p.Value()
		require.NoError(t, err)
		require.Equal(t, "foo", val)
	})

	t.Run("no value", func(t *testing.T) {
		p := Parameter{Type: StringParameterType}

		_, err := p.Value()
		require.Error(t, err)
	})
}

func float64Ptr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}
