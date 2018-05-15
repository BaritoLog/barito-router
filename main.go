package main

import "github.com/BaritoLog/barito-router/router"

func main() {
	r := router.NewRouter(":8080")
	r.Server().ListenAndServe()
}
