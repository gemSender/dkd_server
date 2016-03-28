package game
import (
	"../messages/proto_files"
	math1 "../utility/math"
	b2d "../box2d"
	"../physis/b2Math"
	"log"
)
type PlayerState struct {
	Level int32
	Exp int32
	Index int32
	MaxHp float32
	CurrentHp float32
	Id string
	X float32
	Y float32
	ColorIndex int32
	MsgChan chan messages.GenReplyMsg
	lastPosSyncTime int64
}

func (player *PlayerState) MoveTo (x float32, y float32){
	player.X = x;
	player.Y = y;
}

func (player *PlayerState) StartPath (sx float32, sy float32) {
	player.X = sx;
	player.Y = sy;
}

func (player *PlayerState) GotDamage(damage float32) {
	player.CurrentHp -= damage
	if player.CurrentHp <= 0 {
		player.CurrentHp = 0
		player.Die()
	}
}

func (player *PlayerState) Die() {
	log.Println("player ", player.Index, ": I 'm Dead")
}

func (player *PlayerState) IsDead() bool{
	return player.CurrentHp <= 0
}

func (player *PlayerState) GetCurrentHp() float32 {
	return player.CurrentHp
}

func (player *PlayerState) GetType() ActorType {
	return Player
}

func (player *PlayerState) Shoot(dir math1.Vec2) {
	bulletDef := b2d.NewB2BodyDef()
	bulletDef.SetPosition(b2Math.Vec2{player.X, player.Y})
	bullet := world_instance.phxWorld.CreateBody(bulletDef)
	bulletShape := b2d.NewB2CircleShape()
	bulletShape.SetRadius(1.0)
	bulletFixture := b2d.NewB2FixtureDef()
	bulletFixture.SetIsSensor(true)
	bulletFixture.SetShape(bulletShape)
	bullet.CreateFixture(bulletFixture)
	bullet.SetBullet(true)
	bullet.SetLinearVelocity(b2Math.Vec2(math1.Vec2Mul(dir.Normalized(), 10)))
}