# Changelog

## [0.3.0](https://github.com/damacus/freeagent-cli/compare/v0.2.0...v0.3.0) (2026-03-14)


### Features

* typed client refactor + bills and tasks commands ([#3](https://github.com/damacus/freeagent-cli/issues/3)) ([de561f4](https://github.com/damacus/freeagent-cli/commit/de561f4d248fb5d4cdbbc16b206a1fc2533c5207))

## [0.2.0](https://github.com/damacus/freeagent-cli/compare/v0.1.0...v0.2.0) (2026-03-13)


### Features

* add GitHub CI/CD pipeline with release-please and goreleaser ([910da6b](https://github.com/damacus/freeagent-cli/commit/910da6b89bc5bf39a145a1021d80886d0472d4e6))


### Bug Fixes

* make token refresh thread-safe with sync.Mutex ([cb6b980](https://github.com/damacus/freeagent-cli/commit/cb6b980cc08198a36f1d81238f68f5701386ebd8))
* resolve security, concurrency and reliability issues ([b9204b5](https://github.com/damacus/freeagent-cli/commit/b9204b531c490ba3c0e4c5af35cdd35e65c2cc70))


### Performance Improvements

* reuse HTTP client and parallelise bank approval requests ([f4e1512](https://github.com/damacus/freeagent-cli/commit/f4e15126e33d530196fab98f7d9bb69a22ef1fda))
