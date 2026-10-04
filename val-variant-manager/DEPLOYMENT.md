# Intelligent Agent System: Manual Deployment Guide

This document explains the current **manual deployment flow** for the _Intelligent Agent System Runner_.

At this stage, deployment is not fully automated. The developer builds and publishes a release artifact to the GitLab Package Registry, then the target host downloads, extracts, configures, and runs the binary as a `systemd` service.

### Deployment Flow

```text
Developer machine
  └─ build hardened variant
  └─ package artifact
  └─ publish artifact to GitLab Package Registry

Target host
  └─ download published artifact
  └─ extract binary
  └─ create runtime directories
  └─ install mapping data
  └─ configure environment file
  └─ configure systemd unit
  └─ configure CAN / serial permissions
  └─ start service with systemd
```

---

### Requirements

**Developer machine:**

- Go `>= 1.24`
- GNU Make
- Git
- C compiler for CGO builds
- `readelf`, usually provided by `binutils`
- GitLab access token with permission to upload packages
- Access to private IAS Go modules

**Target host:**

- Linux host or SoM target
- `systemd`
- Network access to Kafka, Schema Registry, MongoDB, and GitLab Package Registry
- SocketCAN or virtual CAN interface, depending on deployment mode
- Serial device access for GNSS, if using CAN variants with serial GNSS
- GitLab access token with permission to download packages
- Runtime user and group for IAS, usually `ias:ias`

---

### Supported Variants

The release system detects variants from the `cmd/` directory.

List available variants:

```bash
$ make list-variants
```

Current variants:

```text
ias-beamng
ias-create
ias-f1-24
ias-general
ias-j1939
ias-obd2
```

Common variants:

| Variant | Runtime source | Description |
| --- | --- | --- |
| `ias-general` | CAN + Serial | Registers OBD2, J1939, and CReATE protocols. |
| `ias-obd2` | CAN + Serial | Registers only OBD2. |
| `ias-j1939` | CAN + Serial | Registers only J1939. |
| `ias-create` | CAN + Serial | Registers only CReATE. |
| `ias-beamng` | UDP | Registers BeamNG OutGauge. Default UDP port: `4444`. |
| `ias-f1-24` | UDP | Registers F1 24 telemetry. Default UDP port: `20777`. |

---

## 1. Build and Publish from Developer Machine

### 1.1 Configure Release Variables

The `Makefile` uses these important variables:

| Variable | Default | Description |
| --- | --- | --- |
| `APP_NAME` | `intelligent-agent-system` | Base application name. |
| `VERSION` | `1.0.0` | Release version. |
| `ENVIRONMENT` | `staging` | Runtime metadata embedded into release builds. |
| `VARIANT` | `ias-general` | Variant to build and publish. |
| `GOOS` | `linux` | Release target OS. Currently Linux only. |
| `GOARCH` | `amd64` | Release target architecture. |
| `PACKAGE_NAME` | `$(APP_NAME)` | GitLab Generic Package Registry package name. |
| `GITLAB_URL` | `https://ev-gitlab.mataelang.net` | GitLab base URL. |
| `PROJECT_ID` | `54` | GitLab project ID. |

Artifact naming format:

```text
<intelligent-agent-system>-<variant>-<version>-linux-<arch>.tar.gz
```

Example:

```text
intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz
```

### 1.2 Export GitLab Token

The publish target requires `GITLAB_TOKEN`.

```bash
$ export GITLAB_TOKEN=<gitlab-token>
```

Do not commit or share this token.

### 1.3 Publish One Variant

Example for `ias-general`, version `1.0.0`, Linux AMD64:

```bash
$ make publish VARIANT=ias-general GOARCH=amd64 VERSION=1.0.0 ENVIRONMENT=staging
```

This target runs:

1. `check-env`
2. `release`
3. `upload`

The release step builds a hardened binary, verifies hardening with `readelf`, packages the binary, and uploads the package to GitLab.

Expected local artifact:

```text
bin/intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz
```

Expected package registry URL format:

```text
https://ev-gitlab.mataelang.net/api/v4/projects/54/packages/generic/intelligent-agent-system/<version>/<artifact-file>
```

Example:

```text
https://ev-gitlab.mataelang.net/api/v4/projects/54/packages/generic/intelligent-agent-system/1.0.0/intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz
```

### 1.4 Publish Other Architectures

Build and publish for ARM64:

```bash
$ make publish VARIANT=ias-general GOARCH=arm64 VERSION=1.0.0 ENVIRONMENT=staging
```

Build and publish one variant for all configured architectures:

```bash
$ make publish-arches VARIANT=ias-general VERSION=1.0.0 ENVIRONMENT=staging
```

The default architecture list is:

```text
amd64 arm64
```

---

## 2. Prepare Target Host

The commands below assume this deployment layout:

```text
/opt/ias/
├── bin/
│   └── intelligent-agent-system
└── data/
    └── mappings.json

/etc/ias/
└── ias-general.env

/etc/systemd/system/
└── ias-general.service
```

Adjust names if deploying another variant.

### 2.1 Create Runtime User and Group

```bash
$ sudo groupadd --system ias
$ sudo useradd --system --gid ias --home-dir /opt/ias --shell /usr/sbin/nologin ias
```

If the user or group already exists, keep the existing account.

### 2.2 Create Directories

```bash
$ sudo mkdir -p /opt/ias/bin
$ sudo mkdir -p /opt/ias/data
$ sudo mkdir -p /etc/ias
```

Set ownership and permissions:

```bash
$ sudo chown -R ias:ias /opt/ias
$ sudo chown root:ias /etc/ias
$ sudo chmod 0750 /opt/ias
$ sudo chmod 0750 /opt/ias/bin
$ sudo chmod 0750 /opt/ias/data
$ sudo chmod 0750 /etc/ias
```

---

## 3. Download and Extract Published Artifact

Set deployment variables on the target host:

```bash
$ export GITLAB_TOKEN=<gitlab-token>
$ export VERSION=1.0.0
$ export VARIANT=ias-general
$ export GOARCH=amd64
$ export ARTIFACT=intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz
```

Download the package:

```bash
$ curl --fail --location --show-error \
  --header "PRIVATE-TOKEN: <gitlab-token>" \
  --output /tmp/intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz \
  https://ev-gitlab.mataelang.net/api/v4/projects/54/packages/generic/intelligent-agent-system/1.0.0/intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz
```

Extract into the runtime binary directory:

```bash
$ sudo tar -xzf /tmp/intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz -C /opt/ias/bin
```

The archive contains the variant-specific binary name, for example:

```text
/opt/ias/bin/intelligent-agent-system-ias-general
```

Create or update a stable symlink for the service:

```bash
$ sudo ln -sfn /opt/ias/bin/intelligent-agent-system-ias-general /opt/ias/bin/intelligent-agent-system
$ sudo chown -h ias:ias /opt/ias/bin/intelligent-agent-system
$ sudo chown ias:ias /opt/ias/bin/intelligent-agent-system-ias-general
$ sudo chmod 0750 /opt/ias/bin/intelligent-agent-system-ias-general
```

Verify the binary exists:

```bash
$ ls -l /opt/ias/bin
```

---

## 4. Install Mapping Data

The runner loads mapping data from this path relative to `WorkingDirectory`:

```text
./data/mappings.json
```

With the recommended service working directory, this resolves to:

```text
/opt/ias/data/mappings.json
```

Install the mapping file from the repository or from your release bundle:

```bash
$ sudo cp data/mappings.json /opt/ias/data/mappings.json
$ sudo chown ias:ias /opt/ias/data/mappings.json
$ sudo chmod 0640 /opt/ias/data/mappings.json
```

> Note: the current `Makefile` package target archives the binary only. If the target host does not already have `mappings.json`, deploy it manually alongside the binary.

---

## 5. Configure Environment File

Create the environment file for the selected variant:

```bash
$ sudo editor /etc/ias/ias-general.env
```

Example structure:

```env
APP_ENV=staging

VIN=<vehicle-vin-or-test-id>

KAFKA_BROKER_URL=<broker-host-1>:19092,<broker-host-2>:19094,<broker-host-3>:19096
SCHEMA_REGISTRY_URL=http://<schema-registry-host>:8085

CAN_NETWORK=can
CAN_NETWORK_ADDRESS=vcan0

MONGODB_URI=mongodb://<mongodb-host>:27017
MONGODB_DATABASE=vss_dictionary
MONGODB_COLLECTION=signal_dictionary
MONGODB_AUTH_USER=<mongodb-username>
MONGODB_AUTH_PASSWORD=<mongodb-password>

SERIAL_PORT=/dev/ttyACM0
```

Protect the environment file because it may contain credentials:

```bash
$ sudo chown root:ias /etc/ias/ias-general.env
$ sudo chmod 0640 /etc/ias/ias-general.env
```

Important variables:

| Variable | Description |
| --- | --- |
| `APP_ENV` | Runtime environment label, for example `staging` or `production`. |
| `VIN` | Vehicle identifier used in outgoing telemetry. |
| `KAFKA_BROKER_URL` | Kafka broker list. |
| `SCHEMA_REGISTRY_URL` | Confluent Schema Registry URL. |
| `CAN_NETWORK` | CAN network type, usually `can`. |
| `CAN_NETWORK_ADDRESS` | CAN interface, for example `vcan0` or `can0`. |
| `MONGODB_URI` | MongoDB connection URI. |
| `MONGODB_DATABASE` | MongoDB database for signal dictionary data. |
| `MONGODB_COLLECTION` | MongoDB collection for signal dictionary data. |
| `MONGODB_AUTH_USER` | MongoDB username. |
| `MONGODB_AUTH_PASSWORD` | MongoDB password. |
| `SERIAL_PORT` | Serial GNSS device path, for example `/dev/ttyACM0`. |

---

## 6. Configure CAN and Serial Access

### 6.1 Virtual CAN for Testing

For testing with `vcan0`:

```bash
$ sudo modprobe vcan
$ sudo ip link add dev vcan0 type vcan
$ sudo ip link set up vcan0
```

Verify:

```bash
$ ip link show vcan0
```

If `vcan0` is already created, `ip link add` may return an error. In that case, ensure the interface is up:

```bash
$ sudo ip link set up vcan0
```

### 6.2 Physical CAN Interface

For a physical CAN interface, configure the correct bitrate for your device.

Example:

```bash
$ sudo ip link set can0 type can bitrate 500000
$ sudo ip link set up can0
```

Then set:

```env
CAN_NETWORK_ADDRESS=can0
```

### 6.3 Serial Device Permission

Check the configured serial device:

```bash
$ ls -l /dev/ttyACM0
```

Depending on the distribution, serial devices may belong to `dialout`, `uucp`, or another group. Add the IAS user to the relevant group.

Example for `dialout`:

```bash
$ sudo usermod -aG dialout ias
```

Example for `uucp`:

```bash
$ sudo usermod -aG uucp ias
```

If the service uses strict systemd device policy, also allow the serial device in the unit file:

```ini
DeviceAllow=/dev/ttyACM0 rw
```

Adjust the path to match `SERIAL_PORT`.

---

## 7. Configure systemd Service

Create the service file:

```bash
$ sudo editor /etc/systemd/system/ias-general.service
```

Base unit:

```ini
[Unit]
Description=Intelligent Agent System (IAS) - General CAN
After=network-online.target sys-subsystem-net-devices-vcan0.device
Wants=network-online.target
Requires=sys-subsystem-net-devices-vcan0.device
StartLimitIntervalSec=300
StartLimitBurst=3

[Service]
Type=simple
User=ias
Group=ias
ProtectSystem=strict
ProtectHome=yes
NoNewPrivileges=true

WorkingDirectory=/opt/ias
ExecStart=/opt/ias/bin/intelligent-agent-system
Restart=on-failure
RestartSec=10

MemoryMax=1G
MemoryLow=256M
CPUQuota=80%
TasksMax=256
LimitNOFILE=65536
LimitNPROC=4096

CapabilityBoundingSet=CAP_NET_RAW
AmbientCapabilities=CAP_NET_RAW

PrivateTmp=true
ProtectClock=true
ProtectHostname=true
ProtectKernelLogs=true
ProtectKernelModules=true
ProtectKernelTunables=true
ProtectControlGroups=true

RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_CAN

RestrictNamespaces=true
RestrictRealtime=true
RestrictSUIDSGID=true
LockPersonality=true

SystemCallFilter=~@clock @debug @module @mount @obsolete @raw-io @reboot @swap
UMask=0027

DevicePolicy=closed
DeviceAllow=/dev/null rw
DeviceAllow=/dev/zero rw
DeviceAllow=/dev/urandom r
DeviceAllow=char-socketcan rw
DeviceAllow=/dev/ttyACM0 rw

StandardOutput=journal
StandardError=journal
SyslogIdentifier=ias-general-staging
SyslogFacility=daemon

KillMode=process
KillSignal=SIGTERM
TimeoutStopSec=30

EnvironmentFile=-/etc/ias/ias-general.env
Environment="APP_ENV=staging"

[Install]
WantedBy=multi-user.target
```

Update these fields for each deployment:

| Field | Notes |
| --- | --- |
| `Description` | Match the deployed variant. |
| `After` / `Requires` | Use the correct CAN interface device, for example `vcan0` or `can0`. For UDP game variants, CAN device dependencies may be removed. |
| `ExecStart` | Use the stable symlink or the variant-specific binary path. |
| `DeviceAllow=/dev/ttyACM0 rw` | Match `SERIAL_PORT`. Remove if the variant does not use serial. |
| `SyslogIdentifier` | Include variant and environment, for example `ias-general-staging`. |
| `EnvironmentFile` | Match the deployed variant, for example `/etc/ias/ias-general.env`. |
| `Environment="APP_ENV=staging"` | Match the target environment or rely on the environment file. |

Reload systemd:

```bash
$ sudo systemctl daemon-reload
```

Enable service on boot:

```bash
$ sudo systemctl enable ias-general.service
```

---

## 8. Start and Verify Service

Start the service:

```bash
$ sudo systemctl start ias-general.service
```

Check service status:

```bash
$ sudo systemctl status ias-general.service
```

Follow logs:

```bash
$ sudo journalctl -u ias-general.service -f
```

Expected behavior:

- service starts under user `ias`
- environment loads from `/etc/ias/ias-general.env`
- mapping file loads from `/opt/ias/data/mappings.json`
- CAN or UDP source starts according to the deployed variant
- Kafka producer connects to the configured broker and Schema Registry
- MongoDB mapper store connects if available
- telemetry is published through the IAS pipeline

---

## 9. Smoke Test

### CAN Variant Smoke Test

For `ias-general`, `ias-obd2`, `ias-j1939`, or `ias-create`, send a test CAN frame from another terminal.

OBD2 example:

```bash
$ cansend vcan0 98DAF115#04410C1770000000
```

J1939 example:

```bash
$ cansend vcan0 18F00400#FFFFFFFF401FFFFF
```

CReATE example:

```bash
$ cansend vcan0 98904001#01F4020D75940357
```

Then inspect logs:

```bash
$ sudo journalctl -u ias-general.service -f
```

### Game Variant Smoke Test

For BeamNG:

- configure the game OutGauge destination IP to the IAS host IP
- configure destination port `4444`

For F1 24:

- configure UDP telemetry destination IP to the IAS host IP
- configure destination port `20777`

Then inspect logs:

```bash
$ sudo journalctl -u ias-beamng.service -f
```

or:

```bash
$ sudo journalctl -u ias-f1-24.service -f
```

---

## 10. Upgrade Existing Deployment

Publish the new version from the developer machine:

```bash
$ export GITLAB_TOKEN=<gitlab-token>
$ make publish VARIANT=ias-general GOARCH=amd64 VERSION=1.0.1 ENVIRONMENT=staging
```

On the target host:

```bash
$ sudo systemctl stop ias-general.service
$ curl --fail --location --show-error \
  --header "PRIVATE-TOKEN: <gitlab-token>" \
  --output /tmp/intelligent-agent-system-ias-general-1.0.1-linux-amd64.tar.gz \
  https://ev-gitlab.mataelang.net/api/v4/projects/54/packages/generic/intelligent-agent-system/1.0.1/intelligent-agent-system-ias-general-1.0.1-linux-amd64.tar.gz
$ sudo tar -xzf /tmp/intelligent-agent-system-ias-general-1.0.1-linux-amd64.tar.gz -C /opt/ias/bin
$ sudo ln -sfn /opt/ias/bin/intelligent-agent-system-ias-general /opt/ias/bin/intelligent-agent-system
$ sudo chown -h ias:ias /opt/ias/bin/intelligent-agent-system
$ sudo chown ias:ias /opt/ias/bin/intelligent-agent-system-ias-general
$ sudo chmod 0750 /opt/ias/bin/intelligent-agent-system-ias-general
$ sudo systemctl start ias-general.service
```

Verify:

```bash
$ sudo systemctl status ias-general.service
$ sudo journalctl -u ias-general.service -n 100 --no-pager
```

---

## 11. Rollback

If the latest deployment fails, switch the stable symlink back to a known-good binary.

Keep previous binaries with versioned filenames when possible, for example:

```text
/opt/ias/bin/intelligent-agent-system-ias-general-1.0.0
/opt/ias/bin/intelligent-agent-system-ias-general-1.0.1
/opt/ias/bin/intelligent-agent-system -> /opt/ias/bin/intelligent-agent-system-ias-general-1.0.0
```

Rollback example:

```bash
$ sudo systemctl stop ias-general.service
$ sudo ln -sfn /opt/ias/bin/intelligent-agent-system-ias-general-1.0.0 /opt/ias/bin/intelligent-agent-system
$ sudo chown -h ias:ias /opt/ias/bin/intelligent-agent-system
$ sudo systemctl start ias-general.service
```

Verify:

```bash
$ sudo systemctl status ias-general.service
$ sudo journalctl -u ias-general.service -n 100 --no-pager
```

---

## 12. Troubleshooting

### Package Download Fails

Check:

- `GITLAB_TOKEN` has package read permission
- `VERSION`, `VARIANT`, and `GOARCH` match the published artifact
- the GitLab project ID is correct
- the target host can reach `https://ev-gitlab.mataelang.net`

### Service Cannot Read Environment File

Check ownership and permissions:

```bash
$ sudo ls -l /etc/ias/ias-general.env
```

Expected:

```text
-rw-r----- root ias /etc/ias/ias-general.env
```

### Service Cannot Read Mapping File

Check:

```bash
$ sudo ls -l /opt/ias/data/mappings.json
```

Expected:

```text
-rw-r----- ias ias /opt/ias/data/mappings.json
```

### CAN Interface Not Found

Check:

```bash
$ ip link show vcan0
```

For virtual CAN:

```bash
$ sudo modprobe vcan
$ sudo ip link add dev vcan0 type vcan
$ sudo ip link set up vcan0
```

For physical CAN, verify the interface name and bitrate.

### Serial Permission Denied

Check the serial device group:

```bash
$ ls -l /dev/ttyACM0
```

Add the `ias` user to the matching group, then restart the service.

Also check whether the service unit needs an explicit `DeviceAllow` entry for the serial path.

### Kafka or Schema Registry Unavailable

Check from the target host:

```bash
$ nc -vz <kafka-host> <kafka-port>
$ curl --fail http://<schema-registry-host>:8085
```

Then verify `KAFKA_BROKER_URL` and `SCHEMA_REGISTRY_URL` in the environment file.

### MongoDB Unavailable

Check from the target host:

```bash
$ nc -vz <mongodb-host> 27017
```

Then verify:

- `MONGODB_URI`
- `MONGODB_DATABASE`
- `MONGODB_COLLECTION`
- `MONGODB_AUTH_USER`
- `MONGODB_AUTH_PASSWORD`

---

### Notes

- The current deployment process is manual.
- The current package artifact contains the binary only.
- Runtime mapping data must be installed separately at `/opt/ias/data/mappings.json` unless packaging is changed later.
- Environment files may contain credentials and must not be committed.
- The service should run as a dedicated non-login user, not as `root`.
- CAN variants require SocketCAN access and usually `CAP_NET_RAW`.
- Serial GNSS access may require both Linux group membership and systemd `DeviceAllow` rules.
- UDP game variants usually do not need CAN device dependencies in the systemd unit.

<br>

> _Authored by Jordan Himawan._<br>
> _web.jojohimawan.cloud_<br>
>
> _Cyber Security Research Group, C304 - D4 Building._ <br>
> _Politeknik Elektronika Negeri Surabaya._ <br>
>
> Ex scientia, veritas. <br>
> Last change: June 12th, 2026.
