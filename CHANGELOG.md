# Changelog

## [0.4.0-rc.1](https://github.com/BeamFoundry/rays-orm/compare/v0.3.1...v0.4.0-rc.1) (2026-11-27)


### ⚠ BREAKING CHANGES

* **tx:** Services should inject core.TransactionManager to control transactions and context; GormTransactionalManager moved to github.com/BeamFoundry/rays-orm/pkg/tx.

### Features

* Add new Rays Stream type as a return type ([0c60303](https://github.com/BeamFoundry/rays-orm/commit/0c6030335d56c81765eca094e10a39a2c6e7d70c))
* **of:** Update to Go 1.27 ([663b01f](https://github.com/BeamFoundry/rays-orm/commit/663b01f6993a5528972602ec26186e66afaae318))
* **otel:** propagate TransactionManager context and add pkg/contrib/otel ([766972c](https://github.com/BeamFoundry/rays-orm/commit/766972c0b151e055e0c705a52a16956373a40104))
* **tx:** integrate with Rays TransactionManager and support DB type… ([#14](https://github.com/BeamFoundry/rays-orm/issues/14)) ([8eb5609](https://github.com/BeamFoundry/rays-orm/commit/8eb5609910d78a4394d11d3e476195a26ce29598))
* **util:** add OrmQueryFunc helpers package ([11c2eea](https://github.com/BeamFoundry/rays-orm/commit/11c2eea1ee8b29bca06b89c49e29607953763448))


### Bug Fixes

* **release:** fix release-please-config ([8a4d9d9](https://github.com/BeamFoundry/rays-orm/commit/8a4d9d9f92e90d6767a3991f3195317d74ca5c96))
* **util:** normalize pagination and handle zero limit and offset overrides ([798cf41](https://github.com/BeamFoundry/rays-orm/commit/798cf41b75b35eaeed0403c95a7e7ea7ef982e08))


### Miscellaneous Chores

* **release:** prepare version v0.4.0-rc.1 ([7f86a2b](https://github.com/BeamFoundry/rays-orm/commit/7f86a2b915003e64363345aab8e1a03c1a9bd40b))
