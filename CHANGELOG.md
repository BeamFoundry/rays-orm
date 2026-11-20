# Changelog

## [0.4.0](https://github.com/BeamFoundry/rays-orm/compare/v0.3.0...v0.4.0) (2026-11-20)


### ⚠ BREAKING CHANGES

* **tx:** Services should inject core.TransactionManager to control transactions and context; GormTransactionalManager moved to github.com/BeamFoundry/rays-orm/pkg/tx.

### Features

* **tx:** integrate with Rays TransactionManager and support DB type… ([#14](https://github.com/BeamFoundry/rays-orm/issues/14)) ([8eb5609](https://github.com/BeamFoundry/rays-orm/commit/8eb5609910d78a4394d11d3e476195a26ce29598))
* **util:** add OrmQueryFunc helpers package ([11c2eea](https://github.com/BeamFoundry/rays-orm/commit/11c2eea1ee8b29bca06b89c49e29607953763448))


### Bug Fixes

* **util:** normalize pagination and handle zero limit and offset overrides ([798cf41](https://github.com/BeamFoundry/rays-orm/commit/798cf41b75b35eaeed0403c95a7e7ea7ef982e08))


### Miscellaneous Chores

* **release:** prepare version v0.4.0 ([b57f410](https://github.com/BeamFoundry/rays-orm/commit/b57f41052f26396a69b52d22c51fddbba685584c))
