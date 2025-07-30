# Barito Router

![Build Status](https://travis-ci.org/BaritoLog/barito-router.svg?branch=master)

Barito Router is the gateway component that routes incoming requests from external sources to the appropriate components within the Barito ecosystem. It supports both gRPC and REST API configurations.

## Overview

Barito Router consists of two main routing components:

### Producer Router

- **Purpose**: Routes log production requests to the appropriate Barito Flow instances
- **Functionality**:
  - Retrieves application profiles from Barito Market based on request headers
  - Uses profile information to route logs to the correct application group
  - Converts incoming REST requests to protobuf before calling Barito Flow's produce API

### Kibana Router

- **Purpose**: Provides reverse proxy functionality for Kibana access
- **Functionality**:
  - Creates Kibana reverse proxy to serve visualization requests
  - Handles authentication and authorization for Kibana access
  - Currently uses REST API (will migrate to gRPC in future versions)

## Features

- **Service Discovery**: Integrates with Barito Market for application profile management
- **Protocol Translation**: Converts REST API calls to gRPC for internal communication
- **Authentication**: Supports SSO authentication and authorization
- **Monitoring**: New Relic integration for performance monitoring
- **Security**: HMAC JWT-based security for secure communication

## Development Setup

### Option 1: Standard Go Setup

```sh
cd $GOPATH/src/github.com/BaritoLog/barito-router
git clone git@github.com:BaritoLog/barito-router.git

cd barito-router
go build
./barito-router
```

### Option 2: Using go get

```sh
go get github.com/BaritoLog/barito-router
$GOPATH/bin/barito-router
```

### Testing and Code Quality

Run unit tests:

```sh
make test
```

Check for vulnerabilities:

```sh
make vuln
```

Check for dead code:

```sh
make deadcode
```

## Configuration

### Environment Variables

| Name | Description | Default Value |
|---|---|---|
| BARITO_PRODUCER_ROUTER | Address that router listen and serve | :8081 |
| BARITO_KIBANA_ROUTER | Address that kibana router listen and serve | :8082 |
| BARITO_MARKET_URL | URL of market API | `http://localhost:3000` |
| BARITO_VIEWER_URL | URL of viewer/router | `http://localhost:8083` |
| BARITO_MARKET_ACCESS_TOKEN | Access token for market API | - |
| BARITO_PROFILE_API_PATH | API path to get app profile by secret | /api/profile |
| BARITO_PROFILE_API_BY_APP_GROUP_PATH | API path to get app profile by app group secret | /api/profile_by_app_group |
| BARITO_AUTHORIZE_API_PATH | API path to authorization | /api/authorize |
| BARITO_PROFILE_API_BY_CLUSTERNAME_PATH | API path to get app profile by cluster name | /api/v2/profile_by_cluster_name |
| BARITO_NEW_RELIC_APP_NAME | Current app name | barito_router |
| BARITO_NEW_RELIC_LICENSE_KEY | License key for kibana router | - |
| BARITO_NEW_RELIC_ENABLED | Enable New Relic agent communication | false |
| BARITO_ENABLE_SSO | Enable SSO authentication | true |
| BARITO_SSO_REDIRECT_PATH | Path for SSO redirect | /auth/callback |
| BARITO_SSO_CLIENT_ID | Client ID for SSO | - |
| BARITO_SSO_CLIENT_SECRET | Client Secret for SSO | - |
| BARITO_HMAC_JWT_SECRET_STRING | HMAC JWT Secret String | - |
| BARITO_ALLOWED_DOMAINS | Allowed domains for SSO | - |

### API Producer Router

- `POST /ping`

  For sending ping to the server
- `POST /produce_batch`

  For sending log entries to be produced on batch by calling barito-flow ProduceBatch API
- `POST /`

  For sending log entries to be produced individually by calling barito-flow Produce API

### API Kibana Router

- `POST /ping`

  For sending ping to the server
- `POST /logout`

  for logging out kibana server
- `POST /`

  create reverse proxy and serve request
