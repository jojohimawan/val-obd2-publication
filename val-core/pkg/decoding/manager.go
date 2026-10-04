package decoding

import (
	"fmt"
	"log/slog"

	"ev-gitlab.mataelang.net/ev-connect/create-ias/can.git"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/config"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/models"
)

type Manager struct {
	handlers    map[uint32]FrameHandler
	gameHandler PacketHandler
	log         *slog.Logger
}

func NewManager() *Manager {
	return &Manager{
		handlers: make(map[uint32]FrameHandler),
		log:      config.ComponentLogger("decoder"),
	}
}

func (m *Manager) RegisterModule(module ProtocolModule) {
	m.log.Info(fmt.Sprintf("Registering protocol %s", module.Name()))

	for id, handler := range module.Handlers() {
		if _, exists := m.handlers[id]; exists {
			m.log.Warn(fmt.Sprintf("handler for ID 0x%X already exists, overwriting with %s", id, module.Name()))
		}

		m.handlers[id] = handler
	}
}

// NOTE: This module is intentionally minimal and test-focused.
// We only care about one specific telemetry packet (PacketID == 6) for F1 24,
// so packet filtering is done inside Decode rather than via routing.
func (m *Manager) RegisterGameModule(module GameModule) {
	slog.Info(fmt.Sprintf("Registering game protocol %s", module.Name()), "component", "core.decoding")
	m.gameHandler = module.GameHandler()
}

func (m *Manager) Decode(frame can.Frame) ([]*models.DecodedSignal, error) {
	handler, exists := m.handlers[frame.ID]
	if !exists {
		m.log.Warn(fmt.Sprintf("No handler found for ID ID 0x%X, skipping", frame.ID))
		return nil, nil
	}

	return handler(frame)
}

func (m *Manager) DecodeGame(data []byte) ([]*models.DecodedSignal, error) {
	return m.gameHandler(data)
}
