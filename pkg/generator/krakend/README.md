# KrakenD Generator

This generator creates environment-specific KrakenD configs from an OpenAPI spec.

Output files:
- `krakend.dev.json`
- `krakend.stage.json`
- `krakend.prod.json`

Generation fails when no `dev`/`stage`/`prod` server endpoint can be detected.

## Usage

```bash
primecodegen openapi-generate -i openapi.yaml -g krakend -t config -o ./krakend \
  --tpl-prop environments=stage,prod \
  --tpl-prop telemetry.otlp.host=otel-collector
```

In app mode (`primelib.yaml`):

```yaml
presets:
  krakend:
    enabled: true
    properties:
      environments: stage,prod
      cors.allowOrigins: https://app.example.com
```

## Generator configuration

Settings that cannot be derived from the spec are template properties of `openapi-krakend-config`.
Resolution order: default < env var `PRIMECODEGEN_TPL_OPENAPI_KRAKEND_CONFIG_<KEY>` (key uppercased, `.` -> `_`) < `--tpl-prop` / `properties`.

| Key                          | Default          | Description                                                                                             |
|------------------------------|------------------|---------------------------------------------------------------------------------------------------------|
| `environments`               | `dev,stage,prod` | Environments to generate configs for                                                                    |
| `server.port`                | `8080`           | Port the gateway listens on                                                                             |
| `server.timeout`             | `3s`             | Default timeout for the whole request pipeline per endpoint                                             |
| `server.readTimeout`         | `10s`            | Maximum duration for reading the entire request                                                         |
| `server.writeTimeout`        | `10s`            | Maximum duration before timing out writes of the response                                               |
| `server.idleTimeout`         | `10s`            | Maximum keep-alive idle time                                                                            |
| `server.readHeaderTimeout`   | `10s`            | Maximum duration for reading request headers                                                            |
| `server.cacheTTL`            | _(empty)_        | Default `Cache-Control` max-age for all endpoints, empty to disable                                     |
| `debugEndpoint.environments` | `dev`            | Environments that expose `/__debug`                                                                     |
| `echoEndpoint.environments`  | `dev`            | Environments that expose `/__echo`                                                                      |
| `cors.enabled`               | `true`           | Enable `security/cors` (allowed methods are derived from the spec operations)                           |
| `cors.allowOrigins`          | `*`              | Comma-separated allowed origins                                                                         |
| `botDetector.enabled`        | `true`           | Enable `security/bot-detector` (rejects empty user agents)                                              |
| `circuitBreaker.enabled`     | `false`          | Enable QoS defaults for all operations (see `x-krakend-circuit-breaker-enabled`)                        |
| `logging.level`              | `INFO`           | `telemetry/logging` level, empty to disable                                                             |
| `telemetry.prometheus.port`  | `9090`           | OpenTelemetry Prometheus exporter port, `0` to disable                                                  |
| `telemetry.otlp.host`        | _(empty)_        | OTLP collector host, empty to disable                                                                   |
| `telemetry.otlp.port`        | `4317`           | OTLP collector port (gRPC)                                                                              |
| `telemetry.traceSampleRate`  | `0.1`            | Trace sample rate (0-1), only used with OTLP                                                            |
| `logging.format`             | `default`        | `default` or `logstash` (also enables `telemetry/logstash`)                                             |
| `logging.gelfAddress`        | _(empty)_        | Graylog `host:port` for `telemetry/gelf`, empty to disable                                              |
| `backend.returnErrorCode`    | `true`           | `backend/http.return_error_code` for non no-op backends (otherwise KrakenD turns backend errors into an empty 500) |
| `validation.requestBody`     | `false`          | `validation/json-schema` from the spec schema of required JSON request bodies                          |
| `securityHeaders.enabled`    | `true`           | `security/http` with frame deny, nosniff, xss filter and `referrer_policy: same-origin`                 |
| `securityHeaders.stsSeconds` | `0`              | HSTS max-age (`sts_seconds`, incl. subdomains), `0` to disable                                          |
| `router.hideVersionHeader`   | `true`           | `router.hide_version_header`                                                                            |
| `router.trustedProxies`      | _(empty)_        | Trusted proxy CIDRs, enables client IP detection via `X-Forwarded-For` (needed for ip rate limits behind a LB) |
| `router.loggerSkipPaths`     | `/__health`      | Paths excluded from the access log                                                                      |

Anything not covered here can be set with the root `x-krakend.service` extension (see below), which is applied last.

## OpenAPI reference

Everything in the spec that influences the generated config. Standard OpenAPI fields are used as-is, gateway specific behaviour is controlled with the `x-krakend` extension.

### Standard OpenAPI fields

| Field                                                                  | Effect                                                                                                                                                         |
|------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `servers[].url` / `servers[].description`                              | Backend host and base path per environment, see [environment mapping](#environment-endpoint-mapping)                                                           |
| `paths.<path>.servers` / `paths.<path>.<method>.servers`               | Per-operation backend override (operation wins over path). Matched to the environment the same way; a single server without an environment marker applies to all environments |
| `paths.<path>` + method                                                | One endpoint per operation, public path == backend path (unless `x-krakend.path` / `endpoint_prefix` are set)                                                  |
| `parameters` with `in: query`                                          | Added to `input_query_strings` (KrakenD drops everything not listed)                                                                                           |
| `parameters` with `in: header`                                         | Added to `input_headers` (`Content-Type`, `Accept` and `User-Agent` are always forwarded)                                                                      |
| `parameters` with `in: cookie`                                         | Forwards the `Cookie` header                                                                                                                                   |
| `security` (document and operation level) + `components.securitySchemes` | Credentials are forwarded to the backend: `apiKey` → its header / query string / `Cookie`, `http`, `oauth2`, `openIdConnect` → `Authorization`. `security: []` marks the operation as public |
| `security` scopes of `oauth2` / `openIdConnect`                        | Become `auth/validator.scopes` (matcher `all`) when JWT validation is configured via `x-krakend.auth`                                                           |
| `oauth2` `flows.clientCredentials.tokenUrl`                            | `token_url` of `auth/client-credentials` when the gateway fetches backend tokens via `x-krakend.client_credentials`                                            |
| `requestBody` (`required: true`, JSON media type, schema)              | Inlined into `validation/json-schema` when `validation.requestBody` is enabled (circular schemas are skipped)                                                  |
| `responses` (success media type + schema)                              | Encoding selection, see [encoding selection](#encoding-selection-behavior)                                                                                     |

Not used: `deprecated`, tags, response headers, examples. Request bodies are always passed through (and optionally validated).

### Root-level `x-krakend`

| Key                 | Type   | Effect                                                                                                                         |
|---------------------|--------|--------------------------------------------------------------------------------------------------------------------------------|
| `service`           | object | Deep-merged into the generated top-level config (applied last, can set anything from the KrakenD schema)                       |
| `endpoint_defaults` | object | Deep-merged into every endpoint                                                                                                |
| `backend_defaults`  | object | Deep-merged into every backend                                                                                                 |
| `endpoint_prefix`   | string | Prefix for all public endpoint paths, e.g. `/petstore` to host multiple APIs on one gateway. Backend paths are not changed      |
| `auth`              | object | Enables JWT validation: base [`auth/validator`](https://www.krakend.io/docs/authorization/jwt-validation/) config for every operation secured by `oauth2`, `openIdConnect` or `http` bearer. Operation scopes are added automatically, `scopes_key` defaults to `scope` |
| `client_credentials`| object | Gateway fetches backend tokens: base [`auth/client-credentials`](https://www.krakend.io/docs/authorization/client-credentials/) (`client_id`, `client_secret`, `scopes`, `endpoint_params`) for operations secured by an oauth2 client credentials scheme, `token_url` comes from the scheme |
| `revoker`           | object | Service [`auth/revoker`](https://www.krakend.io/docs/authorization/revoking-tokens/)                                          |
| `http_server_plugin`| object | Service [`plugin/http-server`](https://www.krakend.io/docs/extending/http-server-plugins/)                                     |

```yaml
x-krakend:
  endpoint_prefix: /petstore
  auth:
    alg: RS256
    jwk_url: https://auth.example.com/.well-known/jwks.json
    issuer: https://auth.example.com
    audience: [petstore]
    cache: true
  service:
    extra_config:
      telemetry/logging:
        level: WARNING
  endpoint_defaults:
    timeout: 2s
  backend_defaults:
    disable_host_sanitize: false
```

### Operation-level `x-krakend`

Location: `paths.<path>.<method>`.

Typed keys:

| Key                   | Type           | Default                                        | Effect                                                                                                                          |
|-----------------------|----------------|------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------|
| `skip`                | bool           | `false`                                        | Do not expose the operation through the gateway                                                                                 |
| `path`                | string         | operation path                                 | Public endpoint path (the backend `url_pattern` keeps the spec path), path params must match the spec path                       |
| `timeout`             | duration       | `server.timeout`                               | Endpoint `timeout`                                                                                                              |
| `cache_ttl`           | duration       | `server.cacheTTL`                              | Endpoint `cache_ttl` (client `Cache-Control` header)                                                                            |
| `concurrent_calls`    | int            | `1`                                            | Endpoint `concurrent_calls`, parallel identical backend requests, use only for idempotent operations                            |
| `input_headers`       | string[]       | derived                                        | Additional forwarded headers, `["*"]` forwards all                                                                              |
| `input_query_strings` | string[]       | derived                                        | Additional forwarded query strings, `["*"]` forwards all                                                                        |
| `output_encoding`     | string         | derived                                        | Endpoint `output_encoding` (`json`, `json-collection`, `xml`, `string`, `negotiate`, `no-op`, ...)                               |
| `backend_encoding`    | string         | derived                                        | Backend `encoding` (`json`, `safejson`, `xml`, `rss`, `string`, `no-op`)                                                        |
| `auth`                | bool \| object | on for bearer-secured operations with root `auth` | `false` disables JWT validation, `true` forces it, an object is merged into `auth/validator` (e.g. `roles`, `roles_key`, `scopes`) |
| `rate_limit`          | bool \| object | off                                            | Endpoint rate limit [`qos/ratelimit/router`](https://www.krakend.io/docs/endpoints/rate-limit/), default `max_rate: 100`       |
| `proxy_rate_limit`    | bool \| object | `circuitBreaker.enabled`                       | Backend rate limit [`qos/ratelimit/proxy`](https://www.krakend.io/docs/backends/rate-limit/), default `max_rate: 100, capacity: 100, every: 30s` |
| `circuit_breaker`     | bool \| object | `circuitBreaker.enabled`                       | [`qos/circuit-breaker`](https://www.krakend.io/docs/backends/circuit-breaker/), default `interval: 60, timeout: 10, max_errors: 1, log_status_change: true` |
| `http_cache`          | bool \| object | `circuitBreaker.enabled`                       | [`qos/http-cache`](https://www.krakend.io/docs/backends/caching/), default `shared: true`                                       |
| `validation`          | bool \| object | `validation.requestBody`                       | [`validation/json-schema`](https://www.krakend.io/docs/endpoints/json-schema/), an object is merged into the derived request body schema (or used as the schema if none is derived) |
| `backend_http`        | bool \| object | `backend.returnErrorCode` (non no-op)          | [`backend/http`](https://www.krakend.io/docs/backends/detailed-errors/), default `return_error_code: true`, e.g. `{return_error_details: backend_alias}` |
| `client_credentials`  | bool \| object | on with root `client_credentials`              | [`auth/client-credentials`](https://www.krakend.io/docs/authorization/client-credentials/), `false` disables, an object is merged into the root config |

For `bool | object` keys an object enables the feature and is deep-merged into its defaults.

Namespace shortcuts, the value is passed through as-is into the matching `extra_config`:

| Key                         | Level    | Namespace                                                                                    |
|-----------------------------|----------|----------------------------------------------------------------------------------------------|
| `signer`                    | endpoint | [`auth/signer`](https://www.krakend.io/docs/authorization/jwt-signing/)                      |
| `cel`                       | endpoint | [`validation/cel`](https://www.krakend.io/docs/endpoints/common-expression-language-cel/)    |
| `lua_proxy`                 | endpoint | [`modifier/lua-proxy`](https://www.krakend.io/docs/endpoints/lua/)                           |
| `lua_endpoint`              | endpoint | [`modifier/lua-endpoint`](https://www.krakend.io/docs/endpoints/lua/)                        |
| `proxy`                     | endpoint | `proxy` ([`sequential`](https://www.krakend.io/docs/endpoints/sequential-proxy/), [`static`](https://www.krakend.io/docs/endpoints/static-proxy/), [`flatmap_filter`](https://www.krakend.io/docs/backends/flatmap/)) |
| `req_resp_modifier`         | endpoint | [`plugin/req-resp-modifier`](https://www.krakend.io/docs/extending/plugin-modifiers/)        |
| `bot_detector`              | endpoint | [`security/bot-detector`](https://www.krakend.io/docs/throttling/botdetector/)               |
| `backend_cel`               | backend  | [`validation/cel`](https://www.krakend.io/docs/endpoints/common-expression-language-cel/)    |
| `lua_backend`               | backend  | [`modifier/lua-backend`](https://www.krakend.io/docs/endpoints/lua/)                         |
| `martian`                   | backend  | [`modifier/martian`](https://www.krakend.io/docs/backends/martian/)                          |
| `backend_proxy`             | backend  | `proxy` ([`shadow`](https://www.krakend.io/docs/backends/shadow-backends/), `flatmap_filter`) |
| `http_client_plugin`        | backend  | [`plugin/http-client`](https://www.krakend.io/docs/extending/http-client-plugins/)           |
| `backend_req_resp_modifier` | backend  | [`plugin/req-resp-modifier`](https://www.krakend.io/docs/extending/plugin-modifiers/)        |

Raw overrides (escape hatch for anything not covered above):

| Key                     | Type   | Effect                                         |
|-------------------------|--------|------------------------------------------------|
| `endpoint`              | object | Deep-merged into the endpoint                  |
| `endpoint_extra_config` | object | Deep-merged into the endpoint `extra_config`   |
| `backend`               | object | Deep-merged into the backend                   |
| `backend_extra_config`  | object | Deep-merged into the backend `extra_config`    |

```yaml
paths:
  /pets:
    get:
      x-krakend:
        path: /animals
        timeout: 5s
        cache_ttl: 60s
        input_headers: [X-Tenant-Id]
        rate_limit:
          max_rate: 50
          client_max_rate: 5
          strategy: ip
        circuit_breaker:
          max_errors: 5
        auth:
          roles_key: realm_access.roles
          roles_key_is_nested: true
          roles: [pet-reader]
        backend:
          disable_host_sanitize: true
  /internal/reindex:
    post:
      x-krakend:
        skip: true
```

Legacy flag: `x-krakend-circuit-breaker-enabled: true|false` on an operation toggles `http_cache`, `circuit_breaker` and `proxy_rate_limit` together (same as the `circuitBreaker.enabled` generator property, but per operation). The typed keys above take precedence.

### Precedence

Lowest to highest:

1. generated values (spec + generator properties)
2. root `x-krakend.endpoint_defaults` / `backend_defaults`
3. operation typed keys (`timeout`, `cache_ttl`, `output_encoding`, ...) and namespace shortcuts
4. operation raw overrides (`endpoint`, `backend`, `*_extra_config`)

Root `x-krakend.service` is merged into the top-level config before endpoints are processed and must not replace `endpoints`.

Generation fails on invalid `x-krakend` values and when two operations end up on the same method + public path.

## KrakenD CE coverage

How each KrakenD Community Edition feature is produced. "derived" = from the spec, "property" = generator property, "op key" / "root key" = `x-krakend`, "raw" = only via `service` / `endpoint` / `backend` / `*_extra_config`.

| Feature / namespace                                  | Level             | How                                                                    |
|------------------------------------------------------|-------------------|------------------------------------------------------------------------|
| endpoints, backends, `url_pattern`, `host`           | endpoint/backend  | derived (`paths`, `servers`), op key `path`, root key `endpoint_prefix` |
| `input_headers`, `input_query_strings`               | endpoint          | derived (parameters, security), op keys                                |
| `output_encoding`, `encoding`, `is_collection`       | endpoint/backend  | derived (responses), op keys                                           |
| `timeout`, `cache_ttl`, `concurrent_calls`           | service/endpoint  | property (service), op keys                                            |
| port, server timeouts, `debug_endpoint`, `echo_endpoint` | service       | property                                                               |
| `router`                                             | service           | property                                                               |
| `security/cors`                                      | service           | property (methods derived)                                             |
| `security/http`                                      | service           | property                                                               |
| `security/bot-detector`                              | service/endpoint  | property, op key `bot_detector`                                        |
| `auth/validator`                                     | endpoint          | derived (security + scopes) with root key `auth`, op key `auth`        |
| `auth/signer`                                        | endpoint          | op key `signer`                                                        |
| `auth/revoker`                                       | service           | root key `revoker`                                                     |
| `auth/client-credentials`                            | backend           | derived (token url) with root key `client_credentials`, op key         |
| `validation/json-schema`                             | endpoint          | derived (request body) with property, op key `validation`              |
| `validation/cel`                                     | endpoint/backend  | op keys `cel`, `backend_cel`                                           |
| `qos/ratelimit/router`                               | endpoint          | op key `rate_limit`                                                    |
| `qos/ratelimit/proxy`, `qos/circuit-breaker`, `qos/http-cache` | backend | property, op keys                                                    |
| `backend/http`                                       | backend           | property, op key `backend_http`                                        |
| `proxy` (sequential, static, flatmap, shadow)        | endpoint/backend  | op keys `proxy`, `backend_proxy`                                       |
| `modifier/lua-proxy`, `modifier/lua-endpoint`, `modifier/lua-backend` | endpoint/backend | op keys                                                 |
| `modifier/martian`                                   | backend           | op key `martian`                                                       |
| `plugin/http-server`                                 | service           | root key `http_server_plugin`                                          |
| `plugin/http-client`, `plugin/req-resp-modifier`     | endpoint/backend  | op keys                                                                |
| `telemetry/logging`, `telemetry/logstash`, `telemetry/gelf` | service    | property                                                               |
| `telemetry/opentelemetry` (prometheus, otlp)         | service           | property                                                               |
| `telemetry/influx`, `telemetry/metrics`, `telemetry/opencensus` | service | raw (`service.extra_config`), the last two are deprecated in favour of OpenTelemetry |
| `tls`, `client_tls`, `async_agent`, `plugin` (top-level) | service       | raw (`service`)                                                        |
| `backend/amqp/*`, `backend/pubsub/*`, `backend/lambda`, `backend/graphql` | backend | raw (`backend_extra_config`), not expressible in an HTTP OpenAPI spec |
| backend `allow`, `deny`, `mapping`, `group`, `target`, `sd` | backend    | raw (`backend`), only effective with non no-op encodings              |

Enterprise-only features (API keys, basic auth, security policies, websockets, gRPC, SOAP, ...) are not generated, but can still be set raw.

## Environment endpoint mapping

OpenAPI `servers[]` are inspected and classified by URL/description tokens:
- `dev`, `development`
- `stage`, `staging`, `uat`, `preprod`, `pre-production`
- `prod`, `production`, `live`

For each environment, the first matching server is used.

Server URL handling:
- `host` keeps the scheme (example: `https://api.example.com`), KrakenD would otherwise default to `http`
- server path prefix is extracted and prepended to each backend `url_pattern`

## Encoding selection behavior

Per operation:
- JSON + array/list return type -> endpoint `json-collection`, backend `json` with `is_collection: true`
- JSON + object/model/map return type -> endpoint `json`, backend `json`
- Anything else (XML, binary, primitive, unknown) -> endpoint `no-op`, backend `no-op`

`no-op` is the safe fallback and is intentionally used whenever type/content is unclear. Override per operation with `x-krakend.output_encoding` / `backend_encoding`.
