package physis

import (
	"../box2d"
	"unsafe"
)

type CollisionSensitive interface{
	BeginContactCallback(Box2D.B2Contact)
	EndContactCallback(Box2D.B2Contact)
}


type GenContactAction struct {
	BeginContactCallback func(Box2D.B2Contact)
	EndContactCallback func(Box2D.B2Contact)
}

func (this *GenContactAction)  BeginContact(contact Box2D.B2Contact){
	if this.BeginContactCallback != nil {
		this.BeginContactCallback(contact)
	}
}

func (this *GenContactAction) EndContact(contact Box2D.B2Contact) {
	if this.EndContactCallback != nil{
		this.EndContactCallback(contact)
	}
}

func NewGenContactListener(beginContactCallback func(Box2D.B2Contact), endContactCallback func(Box2D.B2Contact)) Box2D.B2ContactListener{
	return Box2D.NewDirectorB2ContactListener(&GenContactAction{
		BeginContactCallback:beginContactCallback,
		EndContactCallback:endContactCallback,
	})
}

type GenB2BodyUserData struct {
	Data CollisionSensitive
}

func NewGenB2BodyUserData(data CollisionSensitive) *GenB2BodyUserData{
	ret := &GenB2BodyUserData{Data:data}
	return ret
}

func (this *GenB2BodyUserData)  GetUIntPtr() uintptr{
	return uintptr(unsafe.Pointer(this))
}

func UintPtrPointerToGenB2BodyUserData(ptr uintptr)  *GenB2BodyUserData{
	return (*GenB2BodyUserData)(unsafe.Pointer(ptr))
}