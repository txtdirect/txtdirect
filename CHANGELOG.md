<!--
# v0.4.0
_2020_
  - [Downloads for v0.4.0](https://txtdirect.org/releases/v0.4.0)
  - [Container build for v0.4.0](https://c.txtdirect.org/txtdirect)
  - [Changelog since v0.3.0](#changes-since-v030)
  - [Documentation for v0.4.0](#documentation-for-v040)

## Documentation for v0.2.0
[Documentation](https://txtdirect.org/docs)

[Examples](https://txtdirect.org/examples)

## Changes since v0.1.0
  - Path based redirects

## Fixes since v0.1.0

-->

# v0.4.0
_2020-01-24_
  - [Downloads for v0.4.0](https://txtdirect.org/releases/v0.4.0)
  - [Container build for v0.4.0](https://c.txtdirect.org/txtdirect)
  - [Changelog since v0.3.0](#changes-since-v030)
  - [Documentation for v0.4.0](#documentation-for-v040)

## Documentation for v0.2.0
[Documentation](https://txtdirect.org/docs)

[Examples](https://txtdirect.org/examples)

## Changes since v0.1.0
  - Separate out `gomods` type into its own Caddy plugin [Gomods](https://gomods.okkur.org)
  - Separate out `torproxy` type into its own Caddy plugin [Torproxy](https://torproxy.okkur.org)
  - New `proxy` type [248](https://github.com/txtdirect/txtdirect/pull/248)
  - New `git` type [289](https://github.com/txtdirect/txtdirect/pull/289)
  - New `gometa` type
  - New `dockerv2` type with pull support only
  - Header support [311](https://github.com/txtdirect/txtdirect/pull/311)
  - Use TXTDirect version in builds [310](https://github.com/txtdirect/txtdirect/pull/310)
  - Add TXT record validator tool [304](https://github.com/txtdirect/txtdirect/pull/304)
  - Add predefined regex mode [296](https://github.com/txtdirect/txtdirect/pull/296)
  - Add path chaining [293](https://github.com/txtdirect/txtdirect/pull/293)
  - Add referer header, if not present [280](https://github.com/txtdirect/txtdirect/pull/280)
  - New e2e test coverage [298](https://github.com/txtdirect/txtdirect/pull/298)
  - Support Caddy v1 [266](https://github.com/txtdirect/txtdirect/pull/266)
  - Add regex capture groups as placeholders [257](https://github.com/txtdirect/txtdirect/pull/257)
  - Add more test coverage
  - Support go modules
  - Add Prometheus metrics
  - Redirect permanently with a timeout [192](https://github.com/txtdirect/txtdirect/pull/192)
  - `path`: Wildcard support [147](https://github.com/txtdirect/txtdirect/pull/147)
  - `path`: Path metric per host [250](https://github.com/txtdirect/txtdirect/pull/250)

## Fixes since v0.1.0
  - Normalize path used for zone lookups
  - Various fixes and refactoring to improve stability

---

# v0.3.0
_2019-09-04_
  - [Downloads for v0.3.0](https://txtdirect.org/releases/v0.3.0)
  - [Container build for v0.3.0](https://c.txtdirect.org/txtdirect)
  - [Changelog since v0.2.0](#changes-since-v020)
  - [Documentation for v0.3.0](#documentation-for-v030)

## Documentation for v0.2.0
[Documentation](https://txtdirect.org/docs)

[Examples](https://txtdirect.org/examples)

## Changes since v0.1.0
Intermediate release. All changes will be listed in v0.4.0

-->

# v0.2.0
_2018-08-07_
  - [Downloads for v0.2.0](https://txtdirect.org/releases/v0.2.0)
  - [Container build for v0.2.0](https://c.txtdirect.org/txtdirect)
  - [Changelog since v0.1.0](#changes-since-v010)
  - [Documentation for v0.2.0](#documentation-for-v020)

## Documentation for v0.2.0
[Documentation](https://txtdirect.org/docs)

[Examples](https://txtdirect.org/examples)

## Changes since v0.1.0
  - Path based redirects
  - Enforce absolute zones on DNS requests
  - Blacklist favicon redirects
  - Blacklist `_internal` gometa paths
  - Custom DNS resolver
  - Refactor fallbacks (root fallback, global fallback etc.)
  - Remove Caddy telemetry
  - Update docs
  - Iterate on spec

# v0.1.0
_2017-10-12_
  - [Downloads for v0.1.0](https://txtdirect.org/releases/v0.1.0)
  - [Container build for v0.1.0](https://c.txtdirect.org/txtdirect)
  - [Changelog since v0.0.0](#changes-since-v000)
  - [Documentation for v0.1.0](#documentation-for-v010)

## Documentation for v0.1.0
[Documentation](https://txtdirect.org/docs)

[Examples](https://txtdirect.org/examples)

## Changes since v0.0.0
  - Host based redirects
  - Go package import vanity url redirects
  - 404 defaulting
  - Optional default redirect to "www." subdomain
  - Optional global redirect
