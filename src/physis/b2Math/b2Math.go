package b2Math

import (
	"../../box2d"
	math1 "../../utility/math"
	"unsafe"
)

type Vec2 math1.Vec2

func (this Vec2) Swigcptr() uintptr{
	return uintptr(unsafe.Pointer(&this))
}

func GetVec2FromSwigcptr(ptr Box2D.SwigcptrB2Vec2)  Vec2{
	return *(*Vec2)(unsafe.Pointer(ptr))
}