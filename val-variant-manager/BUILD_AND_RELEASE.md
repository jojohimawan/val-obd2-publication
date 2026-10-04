# Intelligent Agent System: Build and Release Guide

This document explains how to **compile**, **build**, **harden**, **package**, and **publish** the _Intelligent Agent System Runner_ application.

The runner supports multiple application variants under the `cmd/` directory and Linux multi-architecture release builds.

### Supported Variants

Each variant is represented by one directory under `cmd/`.

Current variants:

```text
cmd/ias-beamng
cmd/ias-create
cmd/ias-f1-24
cmd/ias-general
cmd/ias-j1939
cmd/ias-obd2
```

You can list detected variants with:

```bash
$ make list-variants
```

> Some variants may still have compile errors during development. You can limit build/release commands to selected variants by overriding `VARIANTS`.

Example:

```bash
$ make release-variants VARIANTS="ias-general ias-obd2"
```

### Supported Release Targets

The release Makefile currently targets **Linux only**.

Supported architectures:

```text
linux/amd64
linux/arm64
```

Default target:

```text
linux/amd64
```

The default architecture list is controlled by:

```make
TARGET_ARCHES ?= amd64 arm64
```

---

### Requirements

**Build requirements:**

- Go `>= 1.24`
- GNU Make
- Git
- C compiler for CGO builds
- `readelf`, usually from `binutils`

**Publishing requirements:**

- GitLab access token
- Network access to GitLab Package Registry

**Cross-compilation requirements:**

Because this application uses `confluent-kafka-go`, the release build uses **CGO**.

For cross-architecture Linux builds, install the proper cross C compiler.

For building `linux/arm64` from an `amd64` host:

```bash
$ sudo apt install gcc-aarch64-linux-gnu
```

For building `linux/amd64` from an `arm64` host:

```bash
$ sudo apt install gcc-x86-64-linux-gnu
```

The Makefile uses these compiler variables:

```make
CC_AMD64 ?= x86_64-linux-gnu-gcc
CC_ARM64 ?= aarch64-linux-gnu-gcc
```

Override them if your toolchain names are different.

Example:

```bash
$ make release VARIANT=ias-general GOARCH=arm64 CC_ARM64=aarch64-linux-gnu-gcc
```

---

### Makefile Variables

Common variables:

| Variable | Default | Description |
| --- | --- | --- |
| `APP_NAME` | `intelligent-agent-system` | Base application name |
| `VERSION` | `1.0.0` | Release version |
| `ENVIRONMENT` | `staging` | Runtime environment metadata |
| `VARIANT` | `ias-general` | Variant to build |
| `GOOS` | `linux` | Target OS, fixed to Linux |
| `GOARCH` | `amd64` | Target architecture |
| `TARGET_ARCHES` | `amd64 arm64` | Architectures used by multi-arch targets |
| `VARIANTS` | all `cmd/*` variants | Variants used by multi-variant targets |
| `BUILD_DIR` | `bin` | Build output directory |
| `PACKAGE_NAME` | `$(APP_NAME)` | GitLab package name |
| `GITLAB_URL` | `https://ev-gitlab.mataelang.net` | GitLab base URL |
| `PROJECT_ID` | `54` | GitLab project ID |

Example override:

```bash
$ make release VARIANT=ias-general GOARCH=arm64 VERSION=1.2.0 ENVIRONMENT=production
```

---

### Local Compile / Development Build

For local development, use `make build`.

This compiles one variant for the local machine without release packaging.

Default variant:

```bash
$ make build
```

Equivalent to:

```bash
$ make build VARIANT=ias-general
```

Build a specific variant:

```bash
$ make build VARIANT=ias-obd2
```

Build all detected variants locally:

```bash
$ make build-variants
```

Build only selected variants:

```bash
$ make build-variants VARIANTS="ias-general ias-obd2 ias-j1939"
```

Local build output:

```text
bin/intelligent-agent-system-<variant>
```

Example:

```text
bin/intelligent-agent-system-ias-general
```

---

### Hardened Release Build

Use `make release` to build, verify, and package one variant for one Linux architecture.

Default release:

```bash
$ make release
```

Equivalent to:

```bash
$ make release VARIANT=ias-general GOARCH=amd64
```

Build a specific variant:

```bash
$ make release VARIANT=ias-obd2
```

Build for ARM64:

```bash
$ make release VARIANT=ias-general GOARCH=arm64
```

Set production metadata:

```bash
$ make release VARIANT=ias-general GOARCH=amd64 VERSION=1.2.0 ENVIRONMENT=production
```

Release binary output:

```text
bin/<variant>/linux-<arch>/intelligent-agent-system-<variant>
```

Example:

```text
bin/ias-general/linux-amd64/intelligent-agent-system-ias-general
```

Release archive output:

```text
bin/intelligent-agent-system-<variant>-<version>-linux-<arch>.tar.gz
```

Example:

```text
bin/intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz
```

---

### Binary Hardening

Release builds are compiled with binary hardening enabled.

The Makefile uses:

```make
-trimpath
-buildvcs=true
-buildmode=pie
```

External linker hardening:

```make
-Wl,-z,relro,-z,now
-Wl,-z,noexecstack
-Wl,-z,defs
-pie
```

CGO compiler hardening:

```make
-O2
-D_FORTIFY_SOURCE=3
-fstack-protector-strong
-fPIE
```

The `release` target automatically runs `verify-hardening` before packaging.

Manual verification:

```bash
$ make verify-hardening VARIANT=ias-general GOARCH=amd64
```

Alias:

```bash
$ make checksec VARIANT=ias-general GOARCH=amd64
```

The verification checks:

- Binary is an ELF file
- PIE is enabled
- GNU stack is not executable
- GNU RELRO exists
- Immediate binding / `BIND_NOW` is enabled

---

### Multi-Architecture Release

Build one variant for all configured architectures:

```bash
$ make release-arches VARIANT=ias-general
```

By default this builds:

```text
linux/amd64
linux/arm64
```

You can override the architecture list:

```bash
$ make release-arches VARIANT=ias-general TARGET_ARCHES="amd64"
```

or:

```bash
$ make release-arches VARIANT=ias-general TARGET_ARCHES="amd64 arm64"
```

Expected artifacts:

```text
bin/intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz
bin/intelligent-agent-system-ias-general-1.0.0-linux-arm64.tar.gz
```

---

### Multi-Variant Release

Build all variants for one architecture:

```bash
$ make release-variants GOARCH=amd64
```

Build selected variants only:

```bash
$ make release-variants GOARCH=amd64 VARIANTS="ias-general ias-obd2"
```

Build selected variants for ARM64:

```bash
$ make release-variants GOARCH=arm64 VARIANTS="ias-general ias-obd2"
```

---

### Full Release Matrix

Build all selected variants for all selected architectures:

```bash
$ make release-all
```

This combines:

```text
VARIANTS      = all detected cmd/* variants
TARGET_ARCHES = amd64 arm64
```

To avoid variants that are still under development, pass `VARIANTS` explicitly:

```bash
$ make release-all VARIANTS="ias-general ias-obd2 ias-j1939"
```

To build only `amd64`:

```bash
$ make release-all VARIANTS="ias-general ias-obd2" TARGET_ARCHES="amd64"
```

To build both `amd64` and `arm64`:

```bash
$ make release-all VARIANTS="ias-general ias-obd2" TARGET_ARCHES="amd64 arm64"
```

---

### Publishing to GitLab Package Registry

Publishing requires `GITLAB_TOKEN`.

Export your token before publishing:

```bash
$ export GITLAB_TOKEN=glpat-xxxx
```

Publish one variant and one architecture:

```bash
$ make publish VARIANT=ias-general GOARCH=amd64
```

Publish one variant for all architectures:

```bash
$ make publish-arches VARIANT=ias-general
```

Publish all selected variants for one architecture:

```bash
$ make publish-variants GOARCH=amd64 VARIANTS="ias-general ias-obd2"
```

Publish all selected variants for all selected architectures:

```bash
$ make publish-all VARIANTS="ias-general ias-obd2" TARGET_ARCHES="amd64 arm64"
```

The upload target publishes artifacts to:

```text
<GITLAB_URL>/api/v4/projects/<PROJECT_ID>/packages/generic/<PACKAGE_NAME>/<VERSION>/<archive-file>
```

Default package registry location:

```text
https://ev-gitlab.mataelang.net/api/v4/projects/54/packages/generic/intelligent-agent-system/<VERSION>/<archive-file>
```

Example uploaded artifact name:

```text
intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz
```

---

### Common Build Commands

**Build default variant locally**

```bash
$ make build
```

**Build one variant locally**

```bash
$ make build VARIANT=ias-obd2
```

**Build all variants locally**

```bash
$ make build-variants
```

**Release default variant for amd64**

```bash
$ make release
```

**Release one variant for ARM64**

```bash
$ make release VARIANT=ias-general GOARCH=arm64
```

**Release one variant for all architectures**

```bash
$ make release-arches VARIANT=ias-general
```

**Release selected variants for amd64**

```bash
$ make release-variants GOARCH=amd64 VARIANTS="ias-general ias-obd2"
```

**Release selected variants for all architectures**

```bash
$ make release-all VARIANTS="ias-general ias-obd2" TARGET_ARCHES="amd64 arm64"
```

**Publish one artifact**

```bash
$ export GITLAB_TOKEN=glpat-xxxx
$ make publish VARIANT=ias-general GOARCH=amd64
```

**Publish selected variants for all architectures**

```bash
$ export GITLAB_TOKEN=glpat-xxxx
$ make publish-all VARIANTS="ias-general ias-obd2" TARGET_ARCHES="amd64 arm64"
```

---

### Artifact Layout

After release builds, artifacts are stored under `bin/`.

Example tree:

```text
bin/
├── ias-general/
│   ├── linux-amd64/
│   │   └── intelligent-agent-system-ias-general
│   └── linux-arm64/
│       └── intelligent-agent-system-ias-general
├── ias-obd2/
│   ├── linux-amd64/
│   │   └── intelligent-agent-system-ias-obd2
│   └── linux-arm64/
│       └── intelligent-agent-system-ias-obd2
├── intelligent-agent-system-ias-general-1.0.0-linux-amd64.tar.gz
├── intelligent-agent-system-ias-general-1.0.0-linux-arm64.tar.gz
├── intelligent-agent-system-ias-obd2-1.0.0-linux-amd64.tar.gz
└── intelligent-agent-system-ias-obd2-1.0.0-linux-arm64.tar.gz
```

---

### Cleaning Build Artifacts

Remove all generated binaries and archives:

```bash
$ make clean
```

This removes:

```text
bin/
```

---

### Troubleshooting

**1. Unknown variant**

If you see:

```text
ERROR: unknown variant: <variant>
```

Check available variants:

```bash
$ make list-variants
```

Then rerun with a valid variant:

```bash
$ make release VARIANT=ias-general
```

**2. Variant build fails**

Some variants may still be under development.

Limit the release command to known-good variants:

```bash
$ make release-all VARIANTS="ias-general"
```

or:

```bash
$ make release-variants VARIANTS="ias-general ias-obd2"
```

**3. Missing `readelf`**

If you see:

```text
ERROR: readelf is required to verify Linux hardening but was not found.
```

Install `binutils`:

```bash
$ sudo apt install binutils
```

Then rerun:

```bash
$ make release VARIANT=ias-general
```

**4. ARM64 cross-compilation fails**

If building `GOARCH=arm64` from an `amd64` host fails, install the ARM64 cross compiler:

```bash
$ sudo apt install gcc-aarch64-linux-gnu
```

Then rerun:

```bash
$ make release VARIANT=ias-general GOARCH=arm64
```

If your compiler has a different name:

```bash
$ make release VARIANT=ias-general GOARCH=arm64 CC_ARM64=/path/to/aarch64-linux-gnu-gcc
```

**5. GitLab token is missing**

If you see:

```text
ERROR: GITLAB_TOKEN is not set!
```

Export the token:

```bash
$ export GITLAB_TOKEN=glpat-xxxx
```

Then rerun the publish command.

---

### Notes

- Release builds are Linux-only for now.
- Release builds use CGO because the runner depends on Kafka integration through `confluent-kafka-go`.
- Cross-architecture builds require a compatible C cross compiler.
- The `release` target always verifies hardening before packaging.
- The `publish` targets build, verify, package, and upload artifacts.
- Use `VARIANTS` to skip variants that are not ready yet.

<br>

> _Authored by Jordan Himawan._<br>
> _web.jojohimawan.cloud_<br>
>
> _Cyber Security Research Group, C304 - D4 Building._ <br>
> _Politeknik Elektronika Negeri Surabaya._ <br>
>
> Ex scientia, veritas. <br>
> Last change: May 28th, 2026.
`
