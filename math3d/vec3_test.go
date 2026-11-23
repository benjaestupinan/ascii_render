package math3d_test

import (
	"math"
	"testing"

	"github.com/benjaestupinan/ascii_render/math3d"
)

// float32 operations leave small rounding errors (for example after a square
// root), so floats are compared with a tolerance instead of ==.
const epsilon = 1e-6

func near(a, b float32) bool {
	return math.Abs(float64(a-b)) < epsilon
}

func nearVec(a, b math3d.Vec3) bool {
	return near(a.X, b.X) && near(a.Y, b.Y) && near(a.Z, b.Z)
}

var (
	x = math3d.Vec3{X: 1}
	y = math3d.Vec3{Y: 1}
	z = math3d.Vec3{Z: 1}
)

func TestAdd(t *testing.T) {
	tests := []struct {
		a, b, want math3d.Vec3
	}{
		{math3d.Vec3{X: 1, Y: 2, Z: 3}, math3d.Vec3{X: 4, Y: 5, Z: 6}, math3d.Vec3{X: 5, Y: 7, Z: 9}},
		{math3d.Vec3{X: 1, Y: 2, Z: 3}, math3d.Vec3{}, math3d.Vec3{X: 1, Y: 2, Z: 3}},
		{math3d.Vec3{X: 1, Y: -2, Z: 3}, math3d.Vec3{X: -1, Y: 2, Z: -3}, math3d.Vec3{}},
	}
	for _, tt := range tests {
		if got := math3d.Add(tt.a, tt.b); got != tt.want {
			t.Errorf("Add(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		a, b, want math3d.Vec3
	}{
		{math3d.Vec3{X: 5, Y: 7, Z: 9}, math3d.Vec3{X: 4, Y: 5, Z: 6}, math3d.Vec3{X: 1, Y: 2, Z: 3}},
		{math3d.Vec3{X: 1, Y: 2, Z: 3}, math3d.Vec3{X: 1, Y: 2, Z: 3}, math3d.Vec3{}},
		// the order matters: Sub(b, a) goes the opposite way of Sub(a, b)
		{math3d.Vec3{X: 1, Y: 2, Z: 3}, math3d.Vec3{X: 5, Y: 7, Z: 9}, math3d.Vec3{X: -4, Y: -5, Z: -6}},
	}
	for _, tt := range tests {
		if got := math3d.Sub(tt.a, tt.b); got != tt.want {
			t.Errorf("Sub(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestScale(t *testing.T) {
	v := math3d.Vec3{X: 1, Y: -2, Z: 3}
	tests := []struct {
		k    float32
		want math3d.Vec3
	}{
		{2, math3d.Vec3{X: 2, Y: -4, Z: 6}},
		{1, v},
		{0, math3d.Vec3{}},
		{-1, math3d.Vec3{X: -1, Y: 2, Z: -3}},
		{0.5, math3d.Vec3{X: 0.5, Y: -1, Z: 1.5}},
	}
	for _, tt := range tests {
		if got := math3d.Scale(v, tt.k); got != tt.want {
			t.Errorf("Scale(%v, %v) = %v, want %v", v, tt.k, got, tt.want)
		}
	}
}

func TestDot(t *testing.T) {
	tests := []struct {
		name string
		a, b math3d.Vec3
		want float32
	}{
		{"general case", math3d.Vec3{X: 1, Y: 2, Z: 3}, math3d.Vec3{X: 4, Y: 5, Z: 6}, 32}, // 4 + 10 + 18
		{"perpendicular is 0", x, y, 0},
		{"same direction is positive", x, math3d.Scale(x, 3), 3},
		{"opposite direction is negative", x, math3d.Scale(x, -2), -2},
		{"with itself is the length squared", math3d.Vec3{X: 2, Y: 3, Z: 6}, math3d.Vec3{X: 2, Y: 3, Z: 6}, 49},
	}
	for _, tt := range tests {
		if got := math3d.Dot(tt.a, tt.b); !near(got, tt.want) {
			t.Errorf("%s: Dot(%v, %v) = %v, want %v", tt.name, tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCross(t *testing.T) {
	tests := []struct {
		name       string
		a, b, want math3d.Vec3
	}{
		// right-hand rule: X × Y = Z, Y × Z = X, Z × X = Y
		{"X × Y", x, y, z},
		{"Y × Z", y, z, x},
		{"Z × X", z, x, y},
		// swapping the order flips the result
		{"Y × X", y, x, math3d.Scale(z, -1)},
		// parallel vectors do not define a plane: the result is zero
		{"parallel", x, math3d.Scale(x, 5), math3d.Vec3{}},
		{"general case", math3d.Vec3{X: 1, Y: 2, Z: 3}, math3d.Vec3{X: 4, Y: 5, Z: 6}, math3d.Vec3{X: -3, Y: 6, Z: -3}},
	}
	for _, tt := range tests {
		if got := math3d.Cross(tt.a, tt.b); !nearVec(got, tt.want) {
			t.Errorf("%s: Cross(%v, %v) = %v, want %v", tt.name, tt.a, tt.b, got, tt.want)
		}
	}
}

// The cross product is perpendicular to both vectors. This is what will be
// used to get the normal of a triangle.
func TestCrossIsPerpendicularToBoth(t *testing.T) {
	a := math3d.Vec3{X: 1, Y: 2, Z: 3}
	b := math3d.Vec3{X: -2, Y: 0.5, Z: 4}
	c := math3d.Cross(a, b)

	if d := math3d.Dot(c, a); !near(d, 0) {
		t.Errorf("a×b · a = %v, want 0", d)
	}
	if d := math3d.Dot(c, b); !near(d, 0) {
		t.Errorf("a×b · b = %v, want 0", d)
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		v    math3d.Vec3
		want float32
	}{
		{math3d.Vec3{}, 0},
		{x, 1},
		{math3d.Vec3{X: 3, Y: 4}, 5},
		{math3d.Vec3{X: 2, Y: 3, Z: 6}, 7},
		{math3d.Vec3{X: -2, Y: -3, Z: -6}, 7},
	}
	for _, tt := range tests {
		if got := math3d.Length(tt.v); !near(got, tt.want) {
			t.Errorf("Length(%v) = %v, want %v", tt.v, got, tt.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		v, want math3d.Vec3
	}{
		{math3d.Vec3{X: 5}, x},
		{math3d.Vec3{X: 3, Y: 4}, math3d.Vec3{X: 0.6, Y: 0.8}},
		{math3d.Vec3{X: 2, Y: 3, Z: 6}, math3d.Vec3{X: 2.0 / 7, Y: 3.0 / 7, Z: 6.0 / 7}},
		{math3d.Vec3{Z: -0.25}, math3d.Scale(z, -1)},
	}
	for _, tt := range tests {
		got := math3d.Normalize(tt.v)
		if !nearVec(got, tt.want) {
			t.Errorf("Normalize(%v) = %v, want %v", tt.v, got, tt.want)
		}
		if l := math3d.Length(got); !near(l, 1) {
			t.Errorf("Length(Normalize(%v)) = %v, want 1", tt.v, l)
		}
	}
}

// The zero vector has no direction. Dividing by its length (0) would give NaN,
// and a NaN would end up spreading to the light of the whole frame, so
// Normalize returns the zero vector instead.
func TestNormalizeZeroVectorReturnsZero(t *testing.T) {
	got := math3d.Normalize(math3d.Vec3{})
	if got != (math3d.Vec3{}) {
		t.Errorf("Normalize(Vec3{}) = %v, want the zero vector", got)
	}
}
