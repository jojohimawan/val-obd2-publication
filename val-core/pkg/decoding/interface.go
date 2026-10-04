package decoding

import (
	"ev-gitlab.mataelang.net/ev-connect/create-ias/can.git"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/models"
)

type FrameHandler func(can.Frame) ([]*models.DecodedSignal, error)
type PacketHandler func([]byte) ([]*models.DecodedSignal, error)

type ProtocolModule interface {
	Handlers() map[uint32]FrameHandler
	Name() string
}

type GameModule interface {
	GameHandler() PacketHandler
	Name() string
}
