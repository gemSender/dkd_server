package game

type ActorType int

const (
	Player = ActorType(1)
	Npc    = ActorType(2)
)

type GameActor interface  {
	GotDamage(float32)
	Die()
	GetCurrentHp() float32
	IsDead() bool
	GetType() ActorType
}
