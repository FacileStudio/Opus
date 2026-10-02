# Changelog

All notable changes to Opus are documented here.

## [2.8.1] - 2026-10-02

### Changed

- The container listens on port `80` again, reverting the `27538` default introduced in 2.8.0.

## [2.8.0] - 2026-10-02

### Changed

- The single container now listens on port `27538` instead of `80`. Point your reverse proxy
  (Traefik/Dokploy) at container port `27538`.

### Fixed

- Helm chart `opus.service` port, Ingress and HTTPRoute backends now default to the port the image
  actually listens on, instead of the decommissioned `5173`.
- The deployment container's `PORT` is now derived from `opus.service.targetPort`, so the service
  target and the listening port stay in sync.
