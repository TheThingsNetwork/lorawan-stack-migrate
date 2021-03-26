# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), with the exception that this project **does not** adhere to Semantic Versioning.

## [Unreleased]

### Added

### Changed

### Fixed

### Removed

### Security

## [v0.4.0] (2021-03-26)

### Changed

- Rate limit RPC calls to TTN v2 to a maximum of 5 calls per second, to respect global TTN v2 rate limits.
- Disable `DevStatusReq` MAC command for devices exported from TTN v2.

### Fixed

- Retry when receiving errors of type ResourceExhausted and Unavailable, with a backoff.

[0.4.0]: https://github.com/TheThingsNetwork/lorawan-stack/compare/v0.4.0...master
