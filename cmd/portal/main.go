package main

import (
	"context"
	"distributed/log"
	"distributed/portal"
	"distributed/registry"
	"distributed/service"
	"fmt"
	stlog "log"
)

func main() {
	err := portal.ImportTemplates()
	if err != nil {
		stlog.Fatal(err) // Fatal is equivalent to Print() followed by a call to os.Exit(1)
	}
	host, port := "localhost", "5000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	r := registry.Registration{
		ServiceName: registry.PortalService,
		ServiceURL:  serviceAddress,
		RequiredServices: []registry.ServiceName{
			registry.LogService,
			registry.GradingService,
		},
		ServiceUpdateUrl: serviceAddress + "/services",
		HeartbeatURL:     serviceAddress + "/heartbeat",
	}

	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		r,
		portal.RegisterHandlers,
	)
	if err != nil {
		stlog.Fatal(err)
	}

	if logProvider, err := registry.GetProvider(registry.LogService); err != nil {
		log.SetClientLogger(logProvider, r.ServiceName)
	}

	<-ctx.Done()
	fmt.Println("Shutting down portal.")
}
