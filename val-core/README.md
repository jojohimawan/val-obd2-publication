# Vehicle Abstraction Layer: Core

The **Core Module** is the foundational backbone of the Fleet Management System. It provides the essential infrastructure, interfaces, and data pipeline machinery required to build a vehicle telemetry agent
This module is designed to be **protocol-agnostic**, meaning it knows *how* to process data but not *what* specific vehicle data looks like.

### Features
- **Standardization:** All data format are standardized according to **Vehicle Signal Specification** before published.
- **Tracking:** 
  - Collects **NMEA** sentences emitted from GPS module through **serial** port.
  - Currently only handles **latitude** and **longitude** in **GNRMC** sentence.
- **Telemetry:**
  - Collects CAN frames through virtual CAN interface.
  - Collects UDP packets from games intended for **testing only** (you can try with F1 24 and BeamNG).

### Key Packages
- **`pkg/config`**: Handles environment variables and configurations.
- **`pkg/pipeline`**: Implements the concurrent stage-based processing loop (Collect -> Decode -> Marshal -> Publish).
- **`pkg/decoding`**: Defines the `Manager` and `ProtocolModule` interfaces for plugin architecture.
- **`pkg/kafka`**: Manages Kafka producers, topic routing, and Schema Registry integration.
- **`pkg/serial`**: Handles connections to GNSS (Serial) and CAN Bus (SocketCAN).
- **`pkg/mapper`**: MongoDB integration for generic signal mapping.

### Installation
```bash
$ go get ev-gitlab.mataelang.net/ev-connect/create-ias/core.git@latest
```

### Usage Example
This library is intended to be imported by a runner application.
```go
import (
    "ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/decoding"
    "ev-gitlab.mataelang.net/ev-connect/create-ias/core.git/pkg/pipeline"
)

func main() {
    // 1. Initialize the Manager
    manager := decoding.NewManager()

    // 2. Register Protocols (from other modules)
    manager.RegisterModule(obd2.New())
    manager.RegisterGameModule(beamng.New())
    
    // 3. Create a runner config instance
    runner  := pipeline.NewRunner(mapper, producer, manager, vin)
    
    // 4. Run the Pipeline
    runner.Run(ctx, serial, vcan)
    
    // 5. Or try the pipeline using a game!
    runner.RunGame(ctx, udp)
}
```
Look for more in [ev-gitlab.../runner.git](https://ev-gitlab.mataelang.net/ev-connect/create-ias/runner.git).
