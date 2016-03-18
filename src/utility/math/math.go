package math

import "math"

type Vec2 struct {
	X float32
	Y float32
}


func VecAdd(v1 Vec2, v2 Vec2)  Vec2{
	return Vec2{X:v1.X + v2.X, Y:v1.Y + v2.Y}
}

func VecDist(v1 Vec2, v2 Vec2)  float32{
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	return float32(math.Sqrt(float64(dx * dx + dy * dy)))
}

func VecDivide(v Vec2, d float32) Vec2 {
	return  Vec2{X:v.X/d, Y:v.Y/d}
}

func Vec2Minus(v1 Vec2, v2 Vec2)  Vec2{
	return Vec2{X:v1.X - v2.X, Y:v1.Y - v2.Y}
}

func Vec2CrossZ(v1 Vec2, v2 Vec2)  float32{
	return  v1.X * v2.Y - v2.X * v1.Y
}

func Vec2Dot(v1 Vec2, v2 Vec2)  float32{
	return  v1.X * v2.X + v1.Y * v2.Y
}

type Line2d struct {
	Start Vec2
	End Vec2
}

func PointInTriangle(A Vec2, B Vec2, C Vec2, P Vec2)  bool {
	AB := Vec2Minus(B, A)
	AC := Vec2Minus(C, A)
	AP := Vec2Minus(P, A)
	return Vec2CrossZ(AB, AC) * Vec2CrossZ(AB, AP) >= 0
}