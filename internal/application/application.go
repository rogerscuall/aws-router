package application

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"gitlab.presidio.com/rgomez/aws-router/aws/awsrouter"
	"gitlab.presidio.com/rgomez/aws-router/ports"
)

type Application struct {
	RouterClient ports.AWSRouter
}

func NewApplication() *Application {
	return &Application{}
}

// Init will load the credentials into the application. If no credentials are found then an error will be returned.
func (a *Application) Init() error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return ErrNoDefaultAuthentication
	}
	a.RouterClient = ec2.NewFromConfig(cfg)
	return nil
}

// UpdateRouting will identify all the TGWs in a region. It will find all the route tables of the TGWs.
// And it will update the routes on each route table.
func (app *Application) UpdateRouting(ctx context.Context) (tgws []*awsrouter.Tgw, err error) {
	tgws, err = awsrouter.GetAllTgws(ctx, app.RouterClient)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Transit Gateways: %w", err)
	}
	for _, tgw := range tgws {
		if tgw.UpdateRouteTables(ctx, app.RouterClient); err != nil {
			return nil, fmt.Errorf("error retrieving Transit Gateway Route Tables: %w", err)
		}
	}
	// Get all routes from all route tables
	for _, tgw := range tgws {
		tgw.UpdateTgwRoutes(ctx, app.RouterClient)
	}

	return tgws, nil
}
