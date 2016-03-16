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