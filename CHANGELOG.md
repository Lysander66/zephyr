# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.7] - 2025-10-15

### Added

- JSON-RPC client timeout support

## [0.7.6] - 2025-08-13

### Added

- add binary protocol packet implementation with fixed-length header

## [0.7.5] - 2025-07-20

### Added

- Huawei Cloud stream URL generator

## [0.7.4] - 2025-05-22

### Added

- tencent RTMP play url

## [0.7.3] - 2025-01-01

### Added

- pkg add zero value fallbacks

## [0.7.2] - 2025-01-01

### Changed

- Refactor URL generator configuration using Option pattern

## [0.7.1] - 2024-10-12

### Added

- URL-based stream type inference as fallback when Content-Type is inconclusive

## [0.7.0] - 2024-10-08

### Fixed

- Replaced cmd.Start() with cmd.Run() to prevent zombie processes

## [0.6.3] - 2024-09-29

### Changed

- Refactor mergeToMP4 for cross-platform compatibility

## [0.6.2] - 2024-09-20

### Changed

- rename `Run` to `Start`

## [0.6.1] - 2024-09-18

### Fixed

- send on closed channel

## [0.6.0] - 2024-09-18

### Added

- Live stream relaying feature
  - Support for pulling FLV live streams
  - Support for pulling HLS live streams
  - Capability to push streams via RTMP
- Implemented `Relayer` struct for managing stream relaying process

### Changed

- Refactored stream handling logic to support multiple input stream types (FLV, HLS)
- Enhanced publisher interface to accommodate RTMP pushing

## [0.5.0] - 2024-08-12

### Added

- New `run` command for executing operations

### Enhanced

- Improved exit status handling, now supports exit status 127

## [0.4.0] - 2024-08-10

### Added

- Implemented a wrapper for the net/http client

## [0.3.0] - 2024-08-10

### Added

- New Stream URL Generator feature

## [0.2.0] - 2024-08-10

### Added

- Introduced Go bindings for aria2

## [0.1.0] - 2023-06-15

### Added

- Initial release with SSH client functionality to execute commands over SSH connections

[0.7.7]: https://github.com/lysander66/zephyr/compare/v0.7.6...v0.7.7
[0.7.6]: https://github.com/lysander66/zephyr/compare/v0.7.5...v0.7.6
[0.7.5]: https://github.com/lysander66/zephyr/compare/v0.7.4...v0.7.5
[0.7.4]: https://github.com/lysander66/zephyr/compare/v0.7.3...v0.7.4
[0.7.3]: https://github.com/lysander66/zephyr/compare/v0.7.2...v0.7.3
[0.7.2]: https://github.com/lysander66/zephyr/compare/v0.7.1...v0.7.2
[0.7.1]: https://github.com/lysander66/zephyr/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/lysander66/zephyr/compare/v0.6.3...v0.7.0
[0.6.3]: https://github.com/lysander66/zephyr/compare/v0.6.2...v0.6.3
[0.6.2]: https://github.com/lysander66/zephyr/compare/v0.6.1...v0.6.2
[0.6.1]: https://github.com/lysander66/zephyr/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/lysander66/zephyr/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/lysander66/zephyr/compare/v0.3.0...v0.5.0
[0.3.0]: https://github.com/lysander66/zephyr/compare/v0.1.0...v0.3.0
[0.1.0]: https://github.com/lysander66/zephyr/releases/tag/v0.1.0
