package math

import (
	gomath "math"
	"math/big"
	"strings"
	"testing"

	"github.com/goccy/googlesqlite/internal/value"
)

func TestUnaryMinusInt64(t *testing.T) {
	t.Parallel()

	got, err := UNARY_MINUS(value.IntValue(7))
	if err != nil {
		t.Fatalf("UNARY_MINUS(7): %v", err)
	}
	if got != value.IntValue(-7) {
		t.Fatalf("UNARY_MINUS(7) = %v, want -7", got)
	}

	_, err = UNARY_MINUS(value.IntValue(gomath.MinInt64))
	if err == nil || !strings.Contains(err.Error(), "int64 overflow: --9223372036854775808") {
		t.Fatalf("UNARY_MINUS(MinInt64) error = %v", err)
	}
}

func TestUnaryMinusFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{name: "finite", in: 3.25, want: -3.25},
		{name: "positive infinity", in: gomath.Inf(1), want: gomath.Inf(-1)},
		{name: "negative infinity", in: gomath.Inf(-1), want: gomath.Inf(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UNARY_MINUS(value.FloatValue(tt.in))
			if err != nil {
				t.Fatalf("UNARY_MINUS(%v): %v", tt.in, err)
			}
			if got != value.FloatValue(tt.want) {
				t.Fatalf("UNARY_MINUS(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}

	got, err := UNARY_MINUS(value.FloatValue(gomath.NaN()))
	if err != nil {
		t.Fatalf("UNARY_MINUS(NaN): %v", err)
	}
	f, err := got.ToFloat64()
	if err != nil || !gomath.IsNaN(f) {
		t.Fatalf("UNARY_MINUS(NaN) = %v, %v; want NaN", got, err)
	}

	got, err = UNARY_MINUS(value.FloatValue(0))
	if err != nil {
		t.Fatalf("UNARY_MINUS(0): %v", err)
	}
	f, err = got.ToFloat64()
	if err != nil || !gomath.Signbit(f) {
		t.Fatalf("UNARY_MINUS(0) = %v, %v; want negative zero", got, err)
	}
}

func TestUnaryMinusNumeric(t *testing.T) {
	t.Parallel()

	const input = "99999999999999999999999999999.999999999"
	v := numericValue(t, input, false)
	got, err := UNARY_MINUS(v)
	if err != nil {
		t.Fatalf("UNARY_MINUS(NUMERIC): %v", err)
	}
	n, ok := got.(*value.NumericValue)
	if !ok {
		t.Fatalf("UNARY_MINUS(NUMERIC) type = %T", got)
	}
	if n.IsBigNumeric {
		t.Fatal("UNARY_MINUS(NUMERIC) returned BIGNUMERIC")
	}
	want := numericValue(t, "-"+input, false)
	if n.Cmp(want.Rat) != 0 {
		t.Fatalf("UNARY_MINUS(NUMERIC) = %s, want %s", n.Rat, want.Rat)
	}
	if v.Sign() <= 0 {
		t.Fatal("UNARY_MINUS mutated its NUMERIC input")
	}
}

func TestUnaryMinusBigNumeric(t *testing.T) {
	t.Parallel()

	const max = "578960446186580977117854925043439539266.34992332820282019728792003956564819967"
	got, err := UNARY_MINUS(numericValue(t, max, true))
	if err != nil {
		t.Fatalf("UNARY_MINUS(BIGNUMERIC max): %v", err)
	}
	n, ok := got.(*value.NumericValue)
	if !ok || !n.IsBigNumeric {
		t.Fatalf("UNARY_MINUS(BIGNUMERIC) = %T, want BIGNUMERIC", got)
	}
	want := numericValue(t, "-"+max, true)
	if n.Cmp(want.Rat) != 0 {
		t.Fatalf("UNARY_MINUS(BIGNUMERIC) = %s, want %s", n.Rat, want.Rat)
	}

	const min = "-578960446186580977117854925043439539266.34992332820282019728792003956564819968"
	_, err = UNARY_MINUS(numericValue(t, min, true))
	wantErr := "BIGNUMERIC overflow: -(" + min + ")"
	if err == nil || err.Error() != wantErr {
		t.Fatalf("UNARY_MINUS(BIGNUMERIC min) error = %v, want %q", err, wantErr)
	}
}

func numericValue(t *testing.T, s string, isBig bool) *value.NumericValue {
	t.Helper()
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		t.Fatalf("invalid test numeric %q", s)
	}
	return &value.NumericValue{Rat: r, IsBigNumeric: isBig}
}
