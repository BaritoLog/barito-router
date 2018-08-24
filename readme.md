# Barito Router
![alt](https://travis-ci.org/BaritoLog/barito-router.svg?branch=master)

Route incoming request from outside to Barito world.

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
|BARITO_PROFILE_API_PATH|api path to get app profile by secret| /api/profile |
|BARITO_AUTHORIZE_API_PATH|api path to authorization| /api/authorize |
|BARITO_PROFILE_API_BY_CLUSTERNAME_PATH|api path to get app profile by cluster name| /api/profile_by_cluster_name |
