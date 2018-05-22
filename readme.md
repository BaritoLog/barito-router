# Barito Router
![alt](https://travis-ci.org/BaritoLog/barito-router.svg?branch=master)

Route incoming request from external to Barito world

## Setup Development

```sh
cd $GOPATH/src
git clone git@github.com:BaritoLog/barito-router.git

cd barito-router
go build
./barito-router
```

### Env

|Name| Description| Default Value |
|---|---|---|
|BARITO_ROUTER_ADDRESS|Address that router listen and serve|:8081|
|BARITO_KIBANA_ROUTER_ADDRESS|Address that kibana router listen and serve|:8082|
|BARITO_ROUTER_MARKET_URL|URL of market API|http://localhost:3000/api/apps|
