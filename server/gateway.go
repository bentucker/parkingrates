package server

import (
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    gw "github.com/bentucker/parkingrates/genproto"
    "google.golang.org/grpc"
    "fmt"
    "net/http"
    "golang.org/x/net/context"
)

func StartGateway(svcport, gwport int) error {
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    mux := runtime.NewServeMux()
    opts := []grpc.DialOption{grpc.WithInsecure()}
    err := gw.RegisterRatesHandlerFromEndpoint(ctx, mux,
        fmt.Sprintf(":%d", svcport), opts)
    if err != nil {
        return err
    }

    return http.ListenAndServe(fmt.Sprintf(":%d", gwport), mux)
}
