# Vehicle Abstraction Layer: Runner

This module is the **Service Runner (Entry Point)** for the **Fleet Management System.** It acts as the "glue" that binds the **Core Library** (`core`) with specific **Protocol Modules** (`obd2`, `j1939`, `create`) to form a complete vehicle telemetry agent.
Designed to run on a **System on Module (SoM)** embedded in a vehicle, it configures the decoding pipeline, initializes infrastructure (Kafka, Serial, CAN), and executes the main event loop.

### Current Features
- **Modular Architecture:** Plugins-based protocol registration (OBD2, J1939, CReATE).
- **Pipeline Concurrency:** Non-blocking stages for reading, decoding, and publishing.
- **Infrastructure Integration:**
  - Reads from Serial (GNSS), CAN Bus (SocketCAN), and UDP.
  - Publishes to Kafka topics with Schema Registry support.
  - Persists generic mappings via MongoDB.
- **Configurable:** Environment-based configuration via `.env`.
- **Graceful Shutdown:** Handles SIGINT/SIGTERM for clean resource cleanup.

### Pre-Setup

**1. Clone the repository**
```bash
$ git clone https://ev-gitlab.mataelang.net/ev-connect/create-ias/runner.git
$ cd runner
```

**2. Configure Environment**
<br>Create a `.env` file in the root directory. You can copy the example:
```bash
$ cp .env.example .env
```
Ensure variables are set correctly.

**3. Install necessary modules**
<br>Core module
```bash
$ go get ev-gitlab.mataelang.net/ev-connect/create-ias/core.git@latest
```
Decoder module
```bash
$ go get ev-gitlab.mataelang.net/ev-connect/create-ias/protocol-module.git@latest
```

**4. Integrate modules**
<br>CAN variant
```go
import (
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/pipeline"

	// Import the Protocols you want to enable
	create "ev-gitlab.mataelang.net/ev-connect/create-ias/protocol-create.git"
	j1939 "ev-gitlab.mataelang.net/ev-connect/create-ias/protocol-j1939.git"
	obd2 "ev-gitlab.mataelang.net/ev-connect/create-ias/protocol-obd2.git"
)

func main() {
	// Initialize decoder manager.
	decoderManager := decoding.NewManager()

	// Register decoder plugins (modules).
	decoderManager.RegisterModule(obd2.New())
	decoderManager.RegisterModule(j1939.New())
	decoderManager.RegisterModule(create.New())

	// Start pipeline.
	runner := pipeline.NewRunner(mapperService, producer, decoderManager, cfg.Vin)
	runner.Run(ctx, serialReader, vcan)
}
````
<br>Game variant
```go
import (
	"ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/pipeline"
	beamng "ev-gitlab.mataelang.net/ev-connect/create-ias/protocol-beamng.git"
)

func main() {
	// Initialize decoder manager.
	decoderManager := decoding.NewManager()

	// Register decoder plugins (modules).
	decoderManager.RegisterGameModule(beamng.New())

	// Initialize runner
	pipeline.NewRunner(mapperService, producer, decoderManager, cfg.Vin)
	runner.RunGame(ctx, udp)
}
````

### Setup

**1. Create a virtual CAN interface (for testing)**
```bash
$ sudo modprobe vcan
$ sudo ip link add dev vcan0 type vcan
$ sudo ip link set up vcan0
```

**2. Verify Kafka and Schema Registry connectivity**
<br> Ensure that:
- Kafka broker is running and accessible.
- Schema Registry is up and reachable.

**3. Run the service**
```bash
# Run directly
$ go run cmd/ias-general/main.go

# Or build binary
$ go build -o dist/ias-general ./cmd/ias-general
```

**4. Test with sending CAN frame**
<br>In new terminal, try these:
```bash
# Test OBD2
$ cansend vcan0 98DAF115#04410C1770000000
```

### Dev Notes
- This repo uses the "Composition Root" pattern. It contains no business logic, only wiring.
- All decoding logic and pipeline machinery live in [ev-gitlab.../core.git](https://ev-gitlab.mataelang.net/ev-connect/create-ias/core.git).
- Vehicle protocols are imported as separate Go modules.
- To disable a protocol (e.g., CReATE), simply comment out the `RegisterModule()` line in `main.go`.

### Requirements
- Go >= 1.24
- Kafka Broker
- Confluent Schema Registry
- Serial Device (e.g. Arduino Uno, GPS module emitting NMEA sentences)
- Virtual CAN Network Interface (e.g. Linux's can-utils)
- Access to private GitLab modules.
