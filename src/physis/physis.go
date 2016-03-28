package physis

import (
	"../box2d"
)

type genContactListener struct {
	BeginContactCallback func(Box2D.B2Contact)
	EndContactCallback func(Box2D.B2Contact)
}

func (this *genContactListener)  BeginContact(contact Box2D.B2Contact){
	if this.BeginContactCallback != nil {
		this.BeginContactCallback(contact)
	}
}

func (this *genContactListener) EndContact(contact Box2D.B2Contact) {
	if this.EndContactCallback != nil{
		this.EndContactCallback(contact)
	}
}

func NewGenContactListener(beginContactCallback func(Box2D.B2Contact), endContactCallback func(Box2D.B2Contact)) Box2D.B2ContactListener{
	return Box2D.NewDirectorB2ContactListener(&genContactListener{
		BeginContactCallback:beginContactCallback,
		EndContactCallback:endContactCallback,
	})
}