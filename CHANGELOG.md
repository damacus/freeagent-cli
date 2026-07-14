# Changelog

## [0.5.3](https://github.com/damacus/freeagent-cli/compare/v0.5.2...v0.5.3) (2026-07-14)


### Bug Fixes

* **deps:** update module github.com/urfave/cli/v3 to v3.10.0 ([#29](https://github.com/damacus/freeagent-cli/issues/29)) ([c84e466](https://github.com/damacus/freeagent-cli/commit/c84e4660013c79c622c5e7928b338b56aad17859))
* **deps:** update module github.com/urfave/cli/v3 to v3.10.1 ([#32](https://github.com/damacus/freeagent-cli/issues/32)) ([5e23e9a](https://github.com/damacus/freeagent-cli/commit/5e23e9aa04aea09726bb11c756f018524f937636))
* **deps:** update module golang.org/x/sync to v0.22.0 ([#34](https://github.com/damacus/freeagent-cli/issues/34)) ([3829926](https://github.com/damacus/freeagent-cli/commit/3829926824275e34f8ba13f4f675c49ad65dfbed))
* validate token profile paths ([#35](https://github.com/damacus/freeagent-cli/issues/35)) ([3b7b6a1](https://github.com/damacus/freeagent-cli/commit/3b7b6a15b0523c6dc7b0490366a4e51802adf2ee))

## [0.5.2](https://github.com/damacus/freeagent-cli/compare/v0.5.1...v0.5.2) (2026-06-10)


### Bug Fixes

* **deps:** update module github.com/urfave/cli/v2 to v3 ([#25](https://github.com/damacus/freeagent-cli/issues/25)) ([078750d](https://github.com/damacus/freeagent-cli/commit/078750dba3c00dac5942ee306dcdad6e01207301))
* **deps:** update module github.com/urfave/cli/v3 to v3.9.0 ([#27](https://github.com/damacus/freeagent-cli/issues/27)) ([59f7e07](https://github.com/damacus/freeagent-cli/commit/59f7e07f8f5e8d93bd067a4cd3d2a1edaa426c0a))
* **deps:** update module golang.org/x/sync to v0.21.0 ([#28](https://github.com/damacus/freeagent-cli/issues/28)) ([bb18f41](https://github.com/damacus/freeagent-cli/commit/bb18f412bbd3e830aebf291405e5ab989c3312f1))
* report module version for go install builds ([a83c1c1](https://github.com/damacus/freeagent-cli/commit/a83c1c134af6e0258a0c02985a965d73a4cec0d7))

## [0.5.1](https://github.com/damacus/freeagent-cli/compare/v0.5.0...v0.5.1) (2026-05-08)


### Bug Fixes

* **deps:** update module github.com/urfave/cli/v2 to v3 ([#21](https://github.com/damacus/freeagent-cli/issues/21)) ([1e3f6fc](https://github.com/damacus/freeagent-cli/commit/1e3f6fce3feac1fef2648b8fd97a99c4229fe9d7))
* **deps:** update module github.com/urfave/cli/v2 to v3 ([#24](https://github.com/damacus/freeagent-cli/issues/24)) ([7c32da2](https://github.com/damacus/freeagent-cli/commit/7c32da249278d54c7d5df7565026e0736388cdf6))

## [0.5.0](https://github.com/damacus/freeagent-cli/compare/v0.4.0...v0.5.0) (2026-04-15)


### Features

* add version command and contacts update subcommand ([#13](https://github.com/damacus/freeagent-cli/issues/13)) ([d3dd6b9](https://github.com/damacus/freeagent-cli/commit/d3dd6b9c1cc843766fa66716638301ab1945c2ba))

## [0.4.0](https://github.com/damacus/freeagent-cli/compare/v0.3.0...v0.4.0) (2026-03-27)


### Features

* add bank review and accounting commands ([#10](https://github.com/damacus/freeagent-cli/issues/10)) ([7c7d925](https://github.com/damacus/freeagent-cli/commit/7c7d925482599a91958f94529bc823c408143286))


### Bug Fixes

* support trailing receipt flags on updates ([#8](https://github.com/damacus/freeagent-cli/issues/8)) ([36d324c](https://github.com/damacus/freeagent-cli/commit/36d324c7b8533f637683d1afaae52e87a54bc122))

## [0.3.0](https://github.com/damacus/freeagent-cli/compare/v0.2.0...v0.3.0) (2026-03-20)


### Features

* add list and get subcommands for bank transactions ([#5](https://github.com/damacus/freeagent-cli/issues/5)) ([512d6bc](https://github.com/damacus/freeagent-cli/commit/512d6bcd46918516027f130fc556200c8dade1c2))
* **capital-assets:** add capital-assets and capital-asset-types commands ([3068998](https://github.com/damacus/freeagent-cli/commit/306899893afb4df644d8d431c0e14262fe2c2845))
* **categories:** add list/get/create/update/delete commands ([ec8668f](https://github.com/damacus/freeagent-cli/commit/ec8668f1e5f61b4aa2aff411cccb84403e518aa2))
* **company:** add get/business-categories/tax-timeline commands ([ebb6907](https://github.com/damacus/freeagent-cli/commit/ebb69071ba8d14a918b86d83e139c7bb143b8a00))
* **credit-note-reconciliations:** add list/get/create/update/delete commands ([a48b981](https://github.com/damacus/freeagent-cli/commit/a48b981042cedd39f43af5e4bdf0884c7d32f2bf))
* **credit-notes:** add list/get/create/update/delete/transition commands ([bdc5094](https://github.com/damacus/freeagent-cli/commit/bdc509424be62552a3ce3e36be943d5015e70085))
* **estimates:** add list/get/create/update/delete/transition commands ([b73bb32](https://github.com/damacus/freeagent-cli/commit/b73bb32516da72386183b411adcfcf59a2c23881))
* **journal-sets:** add list/get/create/delete/opening-balances commands ([a97b1e2](https://github.com/damacus/freeagent-cli/commit/a97b1e26f6f0982ff80990929726d5721d687217))
* **misc:** add email-addresses, cis-bands, cashflow, accounting commands ([5d885b8](https://github.com/damacus/freeagent-cli/commit/5d885b869f24efbec652eca772f91fdf6e0aaa09))
* **models:** add typed structs for all remaining API resources ([6fc11a0](https://github.com/damacus/freeagent-cli/commit/6fc11a0418c2250702efc4a9e98dfc6a908fc2d8))
* **notes:** add list/get/create/update/delete commands ([4d19a88](https://github.com/damacus/freeagent-cli/commit/4d19a884076a96bfdc1ae1536dedaa5a3523ee3e))
* **payroll:** add payroll and payroll-profiles commands ([a5021f7](https://github.com/damacus/freeagent-cli/commit/a5021f70be61a8ebc165bc5e353f6f5c32b458c6))
* **practice:** add account-managers and clients commands ([15291bd](https://github.com/damacus/freeagent-cli/commit/15291bd54db0c6b392000d62d0aeb4f3d2c7efeb))
* **properties:** add list/get/create/update/delete commands ([6bc835e](https://github.com/damacus/freeagent-cli/commit/6bc835eb5ae2248fe8aef21b826ac70bf8527495))
* **read-only:** add recurring-invoices, stock-items, price-list-items commands ([b5d9797](https://github.com/damacus/freeagent-cli/commit/b5d979749fccd9e3d51da2d5fb2fa1edf49eca37))
* **sales-tax-periods:** add list/get/create/update/delete commands ([39046e2](https://github.com/damacus/freeagent-cli/commit/39046e2925a3dd599f9e60805d752998866f007e))
* typed client refactor + bills and tasks commands ([#3](https://github.com/damacus/freeagent-cli/issues/3)) ([de561f4](https://github.com/damacus/freeagent-cli/commit/de561f4d248fb5d4cdbbc16b206a1fc2533c5207))
* **users:** add list/me/get/create/update/delete commands ([a2a6c5e](https://github.com/damacus/freeagent-cli/commit/a2a6c5e1b93da4b6221e06e89a70c1f985244344))

## [0.2.0](https://github.com/damacus/freeagent-cli/compare/v0.1.0...v0.2.0) (2026-03-13)


### Features

* add GitHub CI/CD pipeline with release-please and goreleaser ([910da6b](https://github.com/damacus/freeagent-cli/commit/910da6b89bc5bf39a145a1021d80886d0472d4e6))


### Bug Fixes

* make token refresh thread-safe with sync.Mutex ([cb6b980](https://github.com/damacus/freeagent-cli/commit/cb6b980cc08198a36f1d81238f68f5701386ebd8))
* resolve security, concurrency and reliability issues ([b9204b5](https://github.com/damacus/freeagent-cli/commit/b9204b531c490ba3c0e4c5af35cdd35e65c2cc70))


### Performance Improvements

* reuse HTTP client and parallelise bank approval requests ([f4e1512](https://github.com/damacus/freeagent-cli/commit/f4e15126e33d530196fab98f7d9bb69a22ef1fda))
