package client

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"

	catalogv1 "github.com/Sokol111/ecommerce-catalog-service-api/gen/go/catalog/v1"
	grpcclient "github.com/Sokol111/ecommerce-commons/pkg/grpc/client"
)

// Module wires a native gRPC client for CatalogService.
// Configuration is read from koanf under key "catalog.grpc".
func Module() fx.Option {
	return fx.Module("product-grpc-client",
		fx.Provide(func(k *koanf.Koanf) (grpcclient.Config, error) {
			return grpcclient.LoadConfig(k, "catalog.grpc")
		}, fx.Private),
		fx.Provide(grpcclient.NewConn, fx.Private),
		fx.Provide(catalogv1.NewProductServiceClient),
		fx.Provide(catalogv1.NewAttributeServiceClient),
		fx.Provide(catalogv1.NewCategoryServiceClient),
	)
}
