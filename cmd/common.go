package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.presidio.com/rgomez/aws-router/adapters/db"
	"gitlab.presidio.com/rgomez/aws-router/aws/awsrouter"
	"gitlab.presidio.com/rgomez/aws-router/ports"
)

var dbNamePrefix string

func syncDb() {
	var err error
	defer func() {
		if err != nil {
			cobra.CheckErr(err)
		}
	}()
	dbNamePrefix = viper.GetString("db_name")
	dbNameTgw := fmt.Sprintf("%s_tgw", dbNamePrefix)
	dbNameTgwRouteTable := fmt.Sprintf("%s_tgw_route_table", dbNamePrefix)
	var dbAdapterTgw, dbAdapterTgwRouteTable ports.DbPort
	dbAdapterTgw, err = db.NewAdapter(dbNameTgw)
	dbAdapterTgwRouteTable, err = db.NewAdapter(dbNameTgwRouteTable)
	defer dbAdapterTgw.CloseDbConnection()
	defer dbAdapterTgwRouteTable.CloseDbConnection()
	fmt.Println("Downloading routing information from AWS")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	client := ec2.NewFromConfig(cfg)
	tgws, err := awsrouter.UpdateRouting(context.TODO(), client)
	fmt.Println("Saving routing information to DB")
	for _, tgw := range tgws {
		err = dbAdapterTgw.SetVal(tgw.ID, tgw.Bytes())
		for _, rt := range tgw.RouteTables {
			err = dbAdapterTgwRouteTable.SetVal(rt.ID, rt.Bytes())
		}
	}
}
