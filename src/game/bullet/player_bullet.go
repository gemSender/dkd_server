package bullet

import (
	b2d "../../box2d"
	"../../physis"
)


type PlayerBullet struct {
	damage float32
}

func (this *PlayerBullet) BeginContactCallback(contact b2d.B2Contact, other physis.CollisionSensitive) {

}