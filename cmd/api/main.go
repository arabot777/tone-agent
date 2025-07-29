package main

import (
	"context"
	"math/rand"
	"time"
	_ "tone/agent/docs"
	"tone/agent/internal/api/resource"
	"tone/agent/internal/api/web"
	"tone/agent/pkg/common/app"
)

// @title T-One API
// @version 1.0
// @description This is a server for T-One API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8888
func main() {
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()
	resource.InitResource(ctx)

	a := app.NewApplication(
		app.AfterStop(resource.Close),
		app.Name("evegen-api"))

	a.AddBundle(web.NewRouter())

	a.Run()
}
