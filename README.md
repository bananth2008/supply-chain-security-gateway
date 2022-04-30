# Supply Chain Security Gateway

A reference architecture and proof of concept implementation of a supply chain security gateway with the goal of enforcing sane security policies to an organization's consumption of 3rd party software (dependencies) in its own products.

- [Supply Chain Security Gateway](#supply-chain-security-gateway)
  - [TL;DR](#tldr)
  - [Architecture](#architecture)
    - [Data Plane Flow](#data-plane-flow)
  - [Usage](#usage)
    - [Authentication](#authentication)
  - [Development](#development)
    - [PDP Development](#pdp-development)
    - [Policy Development](#policy-development)
    - [Tap Development](#tap-development)
    - [Debug NATS Messaging](#debug-nats-messaging)

## TL;DR

Ensure git submodules are updated locally

```bash
git submodule update --init --recursive
```

Initialize keys and certificates for mTLS

```bash
sh bootstrap.sh
```

> This will generate root certificate, per service certificates in `pki/`.

Start the services using `docker-compose`

```bash
docker-compose up -d
```

Verify all the services are active

```bash
docker-compose ps
```

Use the gateway using a `demo-client`

```bash
cd demo-clients/java-gradle && ./gradlew assemble --refresh-dependencies
```

At this point, you should see logs generated by gateway and the policy decision service.

```bash
docker-compose logs envoy
docker-compose logs pdp
```

The `gradle` build should fail with an error message indicating a dependency was blocked by the gateway.

```
> Could not resolve all files for configuration ':app:compileClasspath'.
   > Could not resolve org.apache.logging.log4j:log4j:2.16.0.
     Required by:
         project :app
      > Could not resolve org.apache.logging.log4j:log4j:2.16.0.
         > Could not get resource 'http://localhost:10000/maven2/org/apache/logging/log4j/log4j/2.16.0/log4j-2.16.0.pom'.
            > Could not GET 'http://localhost:10000/maven2/org/apache/logging/log4j/log4j/2.16.0/log4j-2.16.0.pom'. Received status code 403 from server: Forbidden
```

> Refer to `policies/example.rego` for the policy that blocked this artefact

## Architecture

![HLD](docs/images/supply-chain-gateway-hld.png)

### Data Plane Flow

![Data Plane Flow](docs/images/data-plane-flow.png)

## Usage

If you are developing on any of the service and want to force re-create the containers with updated image:

```bash
docker-compose up --force-recreate --remove-orphans --build -d
```

### Authentication

No real authentication is supported yet. Following are planned:

* Gateway authentication using pluggable IDP such as Github OIDC
* Upstream repository basic authentication

## Development

### PDP Development

Build and run the PDP using:

```bash
cd services && make
GLOBAL_CONFIG_PATH=../config/global.yml PDP_POLICY_PATH=../policies ./out/pdp-server
```

PDP listens on `0.0.0.0:9000`. To use the host instance of PDP, edit `config/envoy.yml` and set the address of the `ExtAuthZ` plugin to your host network address.

### Policy Development

Policies are written in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/) and evaluated with [Open Policy Agent](https://www.openpolicyagent.org/docs/latest/integration/#integrating-with-the-go-api)

To run policy test cases:

```bash
cd policies && make test
```

* Refer to `policies/example.rego` for policy example
* Policies are load from `./policies` directory

### Tap Development

The *Tap Service* is integrated as a Envoy [ExtProc](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter) filter. This means, it has greater control over Envoy's request processing life-cycle and can make changes if required.

Currently, it is used for publishing events for data collection only but in future may be extended to support other use-cases. Tap service internally implements a handler chain to delegate an Envoy event to its internal handlers. Example:

```go
tapService, err := tap.NewTapService(config, []tap.TapHandlerRegistration{
  tap.NewTapEventPublisherRegistration(),
})
```

To build and use from host:

```bash
cd services && make
GLOBAL_CONFIG_PATH=../config/global.yml ./out/tap-server
```

> To use Tap service from host, edit `envoy.yml` and change address of `ext-proc-tap` cluster.

### Debug NATS Messaging

Start a docker container with `nats` client

```bash
docker run --rm -it \
   --network supply-chain-security-gateway_default \
   -v `pwd`:/workspace \
   synadia/nats-box
```

Subscribe to a subject and receive messages

```bash
GODEBUG=x509ignoreCN=0 nats sub \
   --tlscert=/workspace/pki/tap/server.crt \
   --tlskey=/workspace/pki/tap/server.key \
   --tlsca=/workspace/pki/root.crt \
   --server=tls://nats-server:4222 \
   com.msg.event.upstream.request
```
