package obd2

import (
	"testing"

	"ev-gitlab.mataelang.net/ev-connect/create-ias/can.git"

	gen "ev-gitlab.mataelang.net/ev-connect/create-ias/protocol-obd2.git/gen"
)

func TestModule_Decoding(t *testing.T) {
	// Initialize the module
	module := New()

	// Get the handler for the main OBD2 ID
	handlers := module.Handlers()
	mainID := gen.Messages().OBD2.ID
	handler, exists := handlers[mainID]
	if !exists {
		t.Fatalf("Module did not register handler for ID %x", mainID)
	}

	// Define test cases
	tests := []struct {
		name          string
		setupFrame    func() can.Frame
		expectParam   string
		expectValue   float64
		expectUnit    string
		expectSuccess bool
	}{
		{
			name: "PID 0x0C Engine RPM",
			setupFrame: func() can.Frame {
				data := [8]uint8{
					0x04,             // Length
					0x41,             // Mode (Response)
					0x0C,             // PID (RPM)
					0x0F,             // A
					0xA0,             // B
					0x00, 0x00, 0x00, // Padding
				}

				return can.Frame{
					ID:         mainID,
					IsExtended: true,
					Length:     8,
					Data:       data,
				}
			},
			expectParam:   "S01PID0C_EngineRPM", // This must match the variable name in DBC
			expectValue:   1000.0,
			expectUnit:    "rpm",
			expectSuccess: true,
		},
		{
			name: "PID 0x0D Vehicle Speed",
			setupFrame: func() can.Frame {
				data := [8]uint8{
					0x03, // Length
					0x41, // Mode
					0x0D, // PID
					0x32, // A (50)
					0x00, 0x00, 0x00, 0x00,
				}

				return can.Frame{
					ID:         mainID,
					IsExtended: true,
					Length:     8,
					Data:       data,
				}
			},
			expectParam:   "S01PID0D_VehicleSpeed",
			expectValue:   50.0,
			expectUnit:    "km/h",
			expectSuccess: true,
		},
		{
			name: "Unknown PID (should be ignored)",
			setupFrame: func() can.Frame {
				// Manually craft a frame with a random PID (e.g., 0xFF)
				// ID: 0x18DAF110 (or whatever your DBC uses)
				// Data: [Length, Mode, PID, ...]
				f := can.Frame{
					ID:         mainID,
					Length:     8,
					Data:       [8]uint8{0x03, 0x41, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00},
					IsExtended: true,
				}
				return f
			},
			expectSuccess: false, // We expect nil result, no error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := tt.setupFrame()

			// EXECUTE & VERIFY
			results, err := handler(frame)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tt.expectSuccess {
				if len(results) != 0 {
					t.Errorf("Expected 0 results for unknown PID, got %d", len(results))
				}
				return // Test finished for this case
			}

			if len(results) != 1 {
				t.Fatalf("Expected exactly 1 signal, got %d", len(results))
			}

			// Validate Metadata
			sig := results[0]
			if sig.Source != "OBD2" {
				t.Errorf("Expected Source 'OBD2', got '%s'", sig.Source)
			}
			if sig.Param != tt.expectParam {
				t.Errorf("Expected Param '%s', got '%s'", tt.expectParam, sig.Param)
			}
			if sig.Unit != tt.expectUnit {
				t.Errorf("Expected Unit '%s', got '%s'", tt.expectUnit, sig.Unit)
			}

			// Validate Value (Float comparison)
			val, ok := sig.Value.(float64)
			if !ok {
				t.Fatalf("Result value is not a float64")
			}

			// Allow a tiny margin of error for floating point math
			epsilon := 0.001
			if (val-tt.expectValue) > epsilon || (tt.expectValue-val) > epsilon {
				t.Errorf("Expected Value %f, got %f", tt.expectValue, val)
			}
		})
	}
}
