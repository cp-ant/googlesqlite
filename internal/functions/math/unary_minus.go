package math

import (
	"fmt"
	"math"
	"math/big"

	"github.com/goccy/googlesqlite/internal/value"
)

var (
	bigNumericMin = mustRat("-578960446186580977117854925043439539266.34992332820282019728792003956564819968")
	bigNumericMax = mustRat("578960446186580977117854925043439539266.34992332820282019728792003956564819967")
)

// UNARY_MINUS implements GoogleSQL's unary minus operator for the numeric
// runtime values supported by googlesqlite. The result retains the input type.
func UNARY_MINUS(x value.Value) (value.Value, error) {
	switch v := x.(type) {
	case value.IntValue:
		if v == value.IntValue(math.MinInt64) {
			return nil, fmt.Errorf("int64 overflow: -%d", v)
		}
		return value.IntValue(-v), nil
	case value.FloatValue:
		return value.FloatValue(-v), nil
	case *value.NumericValue:
		return unaryMinusNumeric(v)
	default:
		return nil, fmt.Errorf("unary minus is unsupported for %T", x)
	}
}

func unaryMinusNumeric(v *value.NumericValue) (value.Value, error) {
	if v == nil || v.Rat == nil {
		return nil, fmt.Errorf("unary minus received an invalid numeric value")
	}

	result := new(big.Rat).Neg(v.Rat)
	if v.IsBigNumeric && (result.Cmp(bigNumericMin) < 0 || result.Cmp(bigNumericMax) > 0) {
		input, err := v.ToString()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("BIGNUMERIC overflow: -(%s)", input)
	}
	return &value.NumericValue{Rat: result, IsBigNumeric: v.IsBigNumeric}, nil
}

func mustRat(s string) *big.Rat {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		panic(fmt.Sprintf("invalid numeric bound %q", s))
	}
	return r
}
