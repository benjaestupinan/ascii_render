// Package math3d has the 3D math used to draw scenes: vectors, the operations
// between them, and the objects and scenes built from them.
package math3d

// Vec3 is a point or a direction in 3D space.
type Vec3 struct {
	X, Y, Z float32
}

// Object is a 3D body made of triangles.
type Object struct {
	vertices []Vec3
	triangles [][3]int // each triangle has the indices in vertices of its 3 corners
}

// Scene is the set of objects that are drawn together.
type Scene struct {
	objects []Object
}

// Add returns the sum of v1 and v2, component by component.
func Add(v1 Vec3, v2 Vec3) Vec3

// Sub returns v1 minus v2, component by component. The result is the vector
// that goes from the point v2 to the point v1.
func Sub(v1 Vec3, v2 Vec3) Vec3

// Scale returns v1 with every component multiplied by k. A negative k also
// reverses its direction.
func Scale(v1 Vec3, k float32) Vec3

// Dot returns the dot product of v1 and v2. It is 0 if they are
// perpendicular, positive if they point to the same side and negative if they
// point to opposite sides.
func Dot(v1 Vec3, v2 Vec3) float32

// Cross returns the cross product of v1 and v2: a vector perpendicular to
// both, following the right-hand rule (Cross(X, Y) = Z). Swapping v1 and v2
// reverses the result, and it is the zero vector if they are parallel.
func Cross(v1 Vec3, v2 Vec3) Vec3

// Length returns the length of v.
func Length(v Vec3) float32

// Normalize returns a vector with the direction of v and length 1. If v is
// the zero vector, which has no direction, it returns the zero vector.
func Normalize(v Vec3) Vec3
