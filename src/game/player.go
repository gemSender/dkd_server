package game
import (
	"../messages/proto_files"
)
type PlayerState struct {
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