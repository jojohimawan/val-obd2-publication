package obd2

import (
	"ev-gitlab.mataelang.net/ev-connect/create-ias/can.git"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/can.git/pkg/descriptor"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/decoding"
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/models"

	gen "ev-gitlab.mataelang.net/ev-connect/create-ias/protocol-obd2.git/gen"
)

// PidDecoder defines function signature for a specific PID
type PidDecoder func(can.Frame) (float64, string, string)

// Module maps PIDs (Service 01 Parameter IDs) to specific extraction logic.
type Module struct {
	pidHandlers map[gen.OBD2_S01PID]PidDecoder
}

// New construct the Module and populate it with extractors (decoder).
func New() *Module {
	m := &Module{
		pidHandlers: make(map[gen.OBD2_S01PID]PidDecoder),
	}
	m.registerPIDs()
	return m
}

// Name identifies this module in logs.
func (m *Module) Name() string {
	return "OBD2"
}

// Handlers returns the map that Core will use.
func (m *Module) Handlers() map[uint32]decoding.FrameHandler {
	handlers := make(map[uint32]decoding.FrameHandler)
	mainID := gen.Messages().OBD2.ID

	handlers[mainID] = m.handleFrame
	return handlers
}

// handleFrame is the "router" for this specific protocol.
// It is called by Core when an OBD2 frame arrives.
func (m *Module) handleFrame(f can.Frame) ([]*models.DecodedSignal, error) {
	// Unmarshal the generic OBD2 Frame to get the PID.
	// Returns error when unmarshalling fails.
	msg := gen.NewOBD2()
	if err := msg.UnmarshalFrame(f); err != nil {
		return nil, err
	}

	// Lookup PID in the private map.
	// If none found, simply return nil and ignore it. No need to freak out.
	decode, exists := m.pidHandlers[msg.S01PID()]
	if !exists {
		return nil, nil
	}

	// Execute the specific decoding logic and return the result
	val, name, unit := decode(f)
	return []*models.DecodedSignal{{
		Source: "OBD2",
		Param:  name,
		Value:  val,
		Unit:   unit,
	}}, nil
}

// makeExtractor creates a handler function for a specific signal.
// It captures the 'signal' variable in a closure, so you don't have to write the lookup logic every time.
func makeExtractor(s *descriptor.Signal) PidDecoder {
	return func(f can.Frame) (float64, string, string) {
		// Extract raw value
		var raw float64

		// Check whether a signal is signed or unsigned, then unmarshal it.
		if s.IsSigned {
			raw = float64(s.UnmarshalSigned(f.Data))
		} else {
			raw = float64(s.UnmarshalUnsigned(f.Data))
		}

		// Convert to physical value and return result
		phys := s.ToPhysical(raw)
		return phys, s.Name, s.Unit
	}
}

// registerPIDs populates the lookup map with specific decoding logic.
func (m *Module) registerPIDs() {
	msg := gen.Messages().OBD2

	// You should automate or simplify these codes, try using loop or something else.
	// I register the decoder manually on purpose, according to my PA scope.
	m.pidHandlers[gen.OBD2_S01PID_S01PID0CEngineRPM] = makeExtractor(msg.S01PID0C_EngineRPM)
	m.pidHandlers[gen.OBD2_S01PID_S01PID04CalcEngineLoad] = makeExtractor(msg.S01PID04_CalcEngineLoad)
	m.pidHandlers[gen.OBD2_S01PID_S01PID11ThrottlePosition] = makeExtractor(msg.S01PID11_ThrottlePosition)
	m.pidHandlers[gen.OBD2_S01PID_S01PID05EngineCoolantTemp] = makeExtractor(msg.S01PID05_EngineCoolantTemp)
	m.pidHandlers[gen.OBD2_S01PID_S01PID0FIntakeAirTemperature] = makeExtractor(msg.S01PID0F_IntakeAirTemperature)
	m.pidHandlers[gen.OBD2_S01PID_S01PID10MAFAirFlowRate] = makeExtractor(msg.S01PID10_MAFAirFlowRate)
	m.pidHandlers[gen.OBD2_S01PID_S01PID0DVehicleSpeed] = makeExtractor(msg.S01PID0D_VehicleSpeed)
}
