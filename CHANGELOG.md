# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

These are changes that will probably be included in the next release.

### Added
 * Add benchmarks

## [v0.2.1] - 2021-09-01

### Fixed
 * Fixed panic when Error is first call on a logger (thanks @wutz for report)
 * Updated readme example to match actual output (thanks @seankhliao)


## [v0.2.0] - 2021-07-07

This version supports the logr v1.0 API which contains breaking changes from pre-1.0 versions.

### Changed
 * Support logr v1.0 API 
 * Write ts field before msg in standard output.
 * Use a fixed number of fractional seconds in timestamp.
 
## [v0.1.5] - 2020-11-23

### Added
 * Add DisableLogger and EnableLogger to disable/enable specific named loggers.

### Removed
 * Deprecated Null logger and it will be removed in next version, use logr.Discard instead
 * Deprecated FromContext and NewContext and they will be removed in next version, use logr.FromContext and logr.NewContext 

## [v0.1.4] - 2020-10-11

### Added
 * Add Null logger
 * Add NewNamed convenience function
 * Add Context support

## [v0.1.3] - 2020-09-20

### Changed 
 * Change to deferred loggers that instantiate their configuration on first use.

## [v0.1.2] - 2020-09-13

### Changed 
 * Updated travis configuration to target recent Go versions

## [v0.1.1] - 2020-07-31

### Added
 * New AddCaller and CallerSkip options to add caller information to log messages

## [v0.1.0] - 2020-07-31

Initial release
