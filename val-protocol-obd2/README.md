# Vehicle Abstraction Layer: OBD2 Decoder

This module is a **Protocol Plugin** for the Vehicle Abstraction Layer. It implements the **OBD2 (On-Board Diagnostics)** standard, capable of routing multiplexed PIDs to their specific decoders.

### Features
- **Architecture:** Uses a `PID Router` to handle multiplexing logic.
- **Supported PIDs:**
  - `0x0C` Engine RPM
  - `0x0D` Vehicle Speed
  - `0x05` Engine Coolant Temperature
  - `0x10` MAF Air Flow Rate
  - *(And others as defined in `module.go`)*

### Installation
```bash
$ go get ev-gitlab.mataelang.net/ev-connect/create-ias/protocol-obd2.git
```

### Integration
<br>Register this module into the Core Decoder Manager:.
```go
import obd2 "ev-gitlab.mataelang.net/ev-connect/create-ias/protocol-obd2.git"

func main() {
    manager.RegisterModule(obd2.New())
}
```
