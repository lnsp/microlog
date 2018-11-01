package main

import (
	"fmt"
	"os"

	"github.com/lnsp/microlog/gateway/pkg/tokens"
	"github.com/micro/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "microadm"
	app.Usage = "microlog command line admin toolbox"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "data",
			Usage: "Database path",
			Value: "micro.db",
		},
		cli.StringFlag{
			Name:  "secret",
			Usage: "Secret for token encryption",
			Value: "secret",
		},
	}
	app.HideVersion = true
	app.Commands = []cli.Command{
		{
			Name:     "email",
			Usage:    "generate access tokens for email",
			Category: "token",
			Action:   getEmailToken,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "address",
					Usage: "Email address",
				},
				cli.StringFlag{
					Name:  "user",
					Usage: "User name",
				},
			},
		},
	}
	app.Run(os.Args)
}

func getEmailToken(ctx *cli.Context) {
	/*
	dataSource, err := models.Open(ctx.GlobalString("data"))
	if err != nil {
		fmt.Println("failed to open data source:", err)
		return
	}
	name := ctx.String("user")
	user, err := dataSource.UserByName(name)
	if err != nil {
		fmt.Println("user does not exist:", err)
		return
	}
	*/
	addr := ctx.String("address")
	token, err := tokens.CreateEmailToken([]byte(ctx.GlobalString("secret")), addr, 1, tokens.PurposeConfirmation)
	if err != nil {
		fmt.Println("failed to create email token:", err)
		return
	}
	fmt.Println(token)
}
