# Barito Router
![alt](https://travis-ci.org/BaritoLog/barito-router.svg?branch=master)

Route incoming request from outside to Barito world. Configurable to use gRPC and REST API.
The REST API is run automatically by calling grpc-gateway which is set in barito-flow.

Barito router consists of 2 routers; producer router and Kibana router.

producer router is responsible to retract profile from barito market based on request header,
to be used as basic information to call the right barito flow client, so timber can arrive to 
the right app group. Keep note that the incoming request to barito router producer is REST and 
will be converted to protobuf just before calling barito flow produce API.

Kibana router responsible to create Kibana reserve proxy to serve the request.
For now, Kibana router still uses REST fully as barito market have not yet converged to gRPC.

## Setup Development

```sh
cd $GOPATH/src/github.com/BaritoLog/barito-router
git clone git@github.com:BaritoLog/barito-router.git

cd barito-router
go build
./barito-router
```

or

```sh
go get github.com/BaritoLog/barito-router
$GOPATH/bin/barito-router
```

### Env

|Name| Description| Default Value |
|---|---|---|
|BARITO_PRODUCER_ROUTER|Address that router listen and serve|:8081|
|BARITO_KIBANA_ROUTER|Address that kibana router listen and serve|:8082|
|BARITO_MARKET_URL|URL of market API| http://localhost:3000 |
|BARITO_VIEWER_URL|URL of viewer/router| http://localhost:8083 |
|BARITO_MARKET_ACCESS_TOKEN|Access token for market API| - |
|BARITO_PROFILE_API_PATH|api path to get app profile by secret| /api/profile |
|BARITO_PROFILE_API_BY_APP_GROUP_PATH|api path to get app profile by app group secret| /api/profile_by_app_group |
|BARITO_AUTHORIZE_API_PATH|api path to authorization| /api/authorize |
|BARITO_PROFILE_API_BY_CLUSTERNAME_PATH|api path to get app profile by cluster name| /api/v2/profile_by_cluster_name |
|BARITO_NEW_RELIC_APP_NAME|Current app name|barito_router|
|BARITO_NEW_RELIC_LICENSE_KEY|License key for kibana router| - |
|BARITO_NEW_RELIC_ENABLED|Enabled controls whether the agent will communicate with the New Relic servers and spawn goroutines|false|
|BARITO_ENABLE_SSO|Enable SSO authentication| true |
|BARITO_SSO_REDIRECT_PATH|Path for SSO redirect| /auth/callback |
|BARITO_SSO_CLIENT_ID|Client ID for SSO| - |
|BARITO_SSO_CLIENT_SECRET|Client Secret for SSO| - |
|BARITO_ALLOWED_DOMAINS|Allowed domains for SSO| - |


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
