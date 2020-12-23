# Changelog

0.7.3.2
- add appName label on producer latency metrics

0.7.3.1
- support disable appName label on metrics

0.7.3
- Enable prometheus instrumentation

0.7.2
- Add tracing and enable verbose error log

0.7.1
- Add tracing capabilities
- Add verbose error when produce failed

0.7.0
- Implement opentracing jaeger to router producer and router viewer

0.6.6

- Accept multiple Consul address

0.6.5

- update consul to randomize service from multiple service available
- tidy up code on consul module
- fix indentation on consul module

**0.6.4**

- bugfix bandwidth exceeded error

**0.6.3**

- Add EnvProducerPort variable

**0.6.2**

- Fix cache key (add appName as cache key) on fetchProfileByAppGroupSecret

**0.6.1**

- Hot-fix cache implementation

**0.6.0**

- Support GRPC
- Enable caching fetch profile from baritoMarket and consulAddrs Producer & Kibana

**0.5.5**

- set CAS cookie path as root

**0.5.4**

- add new param: barito market access token and send it to market when calling fetch profile by cluster name
- use v2 API when fetching profile by cluster name from market

**0.5.3**

- Add support for newrelic

**0.5.2**

- Log unsuccessful request and return the error message back to the requester

**0.5.1**

- Bugfix: Fix proper path for getting Profile based on AppGroup secret

**0.5.0**

- Breaking changes: change params to be sent to market for fetching profile
- Add support for using app group secret and name for fetching profile
