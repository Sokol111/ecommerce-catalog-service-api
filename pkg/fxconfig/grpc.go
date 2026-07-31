package fxconfig

import (
	"context"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	catalogv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/go/catalog/v1"
	"github.com/Sokol111/ecommerce-commons/pkg/core/config"
	grpcclient "github.com/Sokol111/ecommerce-commons/pkg/http/grpc/client"
)

// NewGrpcClientsModule wires a native gRPC client for CatalogService.
// Configuration is read from koanf under key "catalog.grpc".
//
// Provided clients are:
// - catalogv1.ProductServiceClient
// - catalogv1.AttributeServiceClient
// - catalogv1.CategoryServiceClient
func NewGrpcClientsModule() fx.Option {
	return fx.Module("catalog-grpc-clients",
		fx.Provide(func(loader *config.Loader) (grpcclient.Config, error) {
			return grpcclient.LoadConfig(loader, "catalog.grpc")
		}, fx.Private),
		fx.Provide(grpcclient.NewGrpcConnWithTokenSource, fx.Private),
		fx.Provide(func(conn *grpc.ClientConn) grpc.ClientConnInterface {
			return conn
		}, fx.Private),
		fx.Provide(catalogv1.NewProductServiceClient),
		fx.Provide(catalogv1.NewAttributeServiceClient),
		fx.Provide(catalogv1.NewCategoryServiceClient),
		fx.Invoke(func(lc fx.Lifecycle, conn *grpc.ClientConn) {
			lc.Append(fx.Hook{
				OnStop: func(context.Context) error {
					return conn.Close()
				},
			})
		}),
	)
}
