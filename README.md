# Modular Vehicle Abstraction Layer

Supporting code and configuration files for the accompanying manuscript. This repository contains the OBD2-focused subset of the original Vehicle Abstraction Layer (VAL) implementation, aligned with the manuscript's scope.

- [val-core/](val-core/): Shared interfaces and telemetry pipeline infrastructure.
- [val-protocol-obd2/](val-protocol-obd2/): OBD2 decoder module.
- [val-variant-manager/](val-variant-manager/): Application entry point connecting the core and OBD2 decoder.
- [mappings.json](mappings.json): OBD2 signal mappings to Vehicle Signal Specification (VSS) paths, units, and types.
- [obd2.dbc](obd2.dbc): OBD2 CAN message definitions; source: [CSS Electronics](https://www.csselectronics.com/pages/obd2-dbc-file).

The Go modules retain dependencies on the original GitLab repositories; building requires access to those dependencies or adjustments to resolve them locally.
