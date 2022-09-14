# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), with the exception that this project **does not** adhere to Semantic Versioning.

## [Unreleased]

### Added

### Changed

### Fixed

### Removed

### Security

## [v0.9.0] (2022-09-13)

- `ttnv3` source to migrate devices between TTS instances

### Added

## [v0.8.1] (2021-10-12)

### Added

- Apple Silicon builds.

## [v0.8.0] (2021-10-12)

### Added

- `--set-eui-as-id` flag to re-map the ID of end devices. Using this flag will set the device ID to the DevEUI (if it has a DevEUI). When combined with the `--dev-id-prefix` flag, the ID will consist of the prefix and the DevEUI.

## [v0.7.0] (2021-09-27)

### Added

- Support optional prefixes to device IDs.

## [v0.6.0] (2021-06-08)

### Added

- `--ttnv2.resets-to-frequency-plan` flag for configuring factory preset frequencies when exporting ABP devices from The Things Network Stack V2. The list of frequencies is inferred from the chosen Frequency Plan.

### Changed

- Upgrade The Things Stack API to version `3.13.0`.
- Upgrade to Go version `1.16`.

### Fixed

- Properly handle files with Windows-style line endings ("\r\n").

## [v0.5.0] (2021-04-13)

### Changed

- Upgrade The Things Stack API to version `3.12.0`. Due to breaking API changes with The Things Stack 3.12, importing devices that were exported with `ttn-lw-migrate` will fail with previous versions of The Things Stack.

### Removed

- Docker images are no longer built for releases.

## [v0.4.0] (2021-03-26)

### Changed

- Rate limit RPC calls to TTN v2 to a maximum of 5 calls per second, to respect global TTN v2 rate limits.
- Disable `DevStatusReq` MAC command for devices exported from TTN v2.

### Fixed

- Retry when receiving errors of type ResourceExhausted and Unavailable, with a backoff.

<!--
NOTE: These links should respect backports. See https://github.com/TheThingsNetwork/lorawan-stack/pull/1444/files#r333379706.
-->

[unreleased]: https://github.com/TheThingsNetwork/lorawan-stack-migrate/v0.7.0...master
[0.7.0]: https://github.com/TheThingsNetwork/lorawan-stack-migrate/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/TheThingsNetwork/lorawan-stack-migrate/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/TheThingsNetwork/lorawan-stack-migrate/compare/v0.4.0...v0.5.0
