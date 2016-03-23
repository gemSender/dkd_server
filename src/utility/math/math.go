package math

import "math"

type Vec2 struct {
	X float32
	Y float32
}

type GenSegment interface {
	GetStartPoint([]Vec2) Vec2
	GetEndPoint([]Vec2) Vec2
}

func VecAdd(v1 Vec2, v2 Vec2)  Vec2{
	return Vec2{X:v1.X + v2.X, Y:v1.Y + v2.Y}
}

func VecDist(v1 Vec2, v2 Vec2)  float32{
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	return float32(math.Sqrt(float64(dx * dx + dy * dy)))
}

func Vec2Mul(v Vec2, m float32) Vec2 {
	return Vec2{X:v.X*m, Y:v.Y*m}
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

type Segment struct {
	Start Vec2
	End Vec2
}

func (this *Segment) GetStartPoint(vertTable []Vec2) Vec2 {
	return this.Start
}

func (this *Segment) GetEndPoint(vertTable []Vec2) Vec2 {
	return this.End
}

func  PointInTriangle(A Vec2, B Vec2, C Vec2, P Vec2) bool{
	return PBetweenABAC(A, B, C, P) && PBetweenABAC(B, C, A, P) && PBetweenABAC(C, A, B, P);
}

func PBetweenABAC(A Vec2, B Vec2, C Vec2, P Vec2)  bool {
	AP := Vec2Minus(P, A)
	PB := Vec2Minus(B, P)
	PC := Vec2Minus(C, P)
	return Vec2CrossZ(AP, PB) * Vec2CrossZ(AP, PC) <= 0
}

func (v Vec2) Magnitude() float32 {
	return float32(math.Sqrt(float64(v.X * v.X + v.Y * v.Y)))
}

func (v Vec2) Normalized()  Vec2{
	return VecDivide(v, v.Magnitude())
}

func SameDir(v1 Vec2, v2 Vec2)  bool{
	return Vec2Dot(v1, v2) > 0 && math.Abs(float64(Vec2CrossZ(v1, v2))) / float64(v1.Magnitude() * v2.Magnitude()) < 0.00001
}

func DistPointToSegment(A Vec2, B Vec2, P Vec2) (float32, Vec2){
	AP := Vec2Minus(P, A)
	AB := Vec2Minus(B, A)
	DotAPAB := Vec2Dot(AP, AB)
	if DotAPAB <= 0{
		return AP.Magnitude(), A
	}
	LenAB := AB.Magnitude()
	if DotAPAB >= LenAB{
		return VecDist(P, B), B
	}
	r := DotAPAB / LenAB
	C := VecAdd(A, Vec2Mul(AB, r))
	return VecDist(C, P), C
}