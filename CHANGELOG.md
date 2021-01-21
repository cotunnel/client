# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.1] - 2021-01-21

### Added
- TCP/UDP tunnel added.
- Host parameter added to tunnels. Your tunnels can connect different host from localhost.

### Changed
- Dependency manager migrated to Go Modules.
- vendor folder removed from the repository.
- WEB tunnel bug fixed.
- Client update function bug fixed.

## [1.0.3] - 2020-07-01

### Changed
- Tunnel log added. Users can see web tunnel requests on cotunnel client console now.

## [1.0.2] - 2020-06-30

### Changed
- An error related to registration and then client update has been fixed.
- Wrong client update console log fixed.

## [1.0.1] - 2020-06-28

### Changed
- Unnecessary registration console log deleted.

## [1.0.0] - 2020-05-16

### Changed
- Versioning updated.

## [183] - 2020-05-14

### Added
- Tunnel connection (the client <-> tunnel server) updated to TLS. 

### Changed
- Register function updated. You don't need to restart the client after registration if you are using cotunnel client without installation script.
