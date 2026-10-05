# IAS — A Modular Vehicle Abstraction Layer for OBD-II Telemetry Integration

Supporting materials for the manuscript:

> **A Modular Vehicle Abstraction Layer for OBD-II Telemetry Integration**
> Ferry Astika Saputra, Jordan Frisay Himawan, Yesta Medya Mahardhika, Arif Basofi, Dadet Pramadihanto
> Department of Informatics and Computer Engineering, Politeknik Elektronika Negeri Surabaya, Surabaya, Indonesia
> Submitted to *Jurnal RESTI (Rekayasa Sistem dan Teknologi Informasi)*, published by Ikatan Ahli Informatika Indonesia (IAII). **Status: under review.**

This repository is published to support the reproducibility requirements of the journal. It contains the source code of the Intelligent Agent System (IAS), the OBD-II DBC configuration, and the signal dictionary mapping used in the reported experiment.

---

## 1. What the manuscript reports

IAS is implemented in Go as a Vehicle Abstraction Layer sitting between a vehicle-data source and Kafka, on the edge side of an electric-vehicle Fleet Management System. It is organised into three components:

| Component | Responsibility |
|---|---|
| **Core** | Shared processing, infrastructure connections, and the decoder interface |
| **Protocol decoder** | Source-specific interpretation rules, implemented in a separate codebase against the core interface |
| **Variant manager** | Configures the core, instantiates and registers the selected decoder, controls pipeline shutdown |

The evaluated processing path is:

```
CAN frames → collector → OBD-II DBC decoding → VSS-oriented mapping → Protobuf payload → Kafka
                                                                                            ↓
                                                     Grafana ← TimescaleDB ← Telemetry Ingester
```

Each stage runs as a Go goroutine, with Go channels carrying data between stages. Signal mapping uses a MongoDB dictionary (source protocol, original signal name, destination path, unit, type), loaded into a local JSON cache at initialisation.

Reported quantitative result: over a 30-minute observation, the source generated **9,018 CAN frames** and the monitoring layer recorded **9,000 signal observations**, giving a signal yield of **0.998 signals/frame**. As stated in the manuscript, this ratio is a descriptive output-to-input measure, not a delivery percentage, decoding-accuracy measure, or latency benchmark.

---

## 2. Repository contents

- [val-core/](val-core/): Shared interfaces and telemetry pipeline infrastructure.
- [val-protocol-obd2/](val-protocol-obd2/): OBD2 decoder module.
- [val-variant-manager/](val-variant-manager/): Application entry point connecting the core and OBD2 decoder.
- [mappings.json](mappings.json): OBD2 signal mappings to Vehicle Signal Specification (VSS) paths, units, and types.
- [obd2.dbc](obd2.dbc): OBD2 CAN message definitions; source: [CSS Electronics](https://www.csselectronics.com/pages/obd2-dbc-file).

The Go modules retain dependencies on the original GitLab repositories; building requires access to those dependencies or adjustments to resolve them locally.

##  3. How to cite

If you use this material, please cite the manuscript.

```bibtex
@article{saputra_val_obd2,
  author  = {Saputra, Ferry Astika and Himawan, Jordan Frisay and
             Mahardhika, Yesta Medya and Basofi, Arif and Pramadihanto, Dadet},
  title   = {A Modular Vehicle Abstraction Layer for {OBD-II} Telemetry Integration},
  journal = {Jurnal RESTI (Rekayasa Sistem dan Teknologi Informasi)},
  note    = {Manuscript under review},
  year    = {2026}
}
```

## 4. Funding

This research was funded by the Ministry of Higher Education, Science, and Technology of the Republic of Indonesia (Kemendiktisaintek) through the Hiliriset Program under the Hilirisasi Inovasi Komersial scheme, Contract No. 1077/PL14/PT/IX/2025 and Contract No. 1467/DST/PL14/PT/IV/2026.

## 5. Contact 
- Ferry Astika Saputra (First Author): ferryas@pens.ac.id
- Jordan Frisay Himawan (Co-Author): himawanjordan@gmail.com

