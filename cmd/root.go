package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/jaehong21/hibiscus/config"
	app "github.com/jaehong21/hibiscus/tviewapp/hibiscus"
	ecrsvc "github.com/jaehong21/hibiscus/tviewapp/hibiscus/services/ecr"
	elbsvc "github.com/jaehong21/hibiscus/tviewapp/hibiscus/services/elb"
	route53svc "github.com/jaehong21/hibiscus/tviewapp/hibiscus/services/route53"
	"github.com/spf13/cobra"
)

var (
	// NOTE: ldflags
	buildVersion   = "unknown"
	buildDate      = "unknown"
	buildOS        = "unknown"
	buildArch      = "unknown"
	buildCommit    = "unknown"
	buildGoVersion = "unknown"
)

var awsProfile string

func init() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.Flags().StringVarP(&awsProfile, "profile", "p", "default", "AWS profile to use")
}

var rootCmd = &cobra.Command{
	Use:   "hibiscus",
	Short: "Hibiscus is a modern terminal UI for AWS console",
	Long: `Hibiscus is a modern terminal UI for AWS console. 
            It is built with tview and cobra.
            It aims to provide a simple and intuitive way to interact with AWS services.`,
	Run: func(cmd *cobra.Command, args []string) {
		newConfig := config.Initialize()
		config.SetAwsProfile(awsProfile)

		factories := []app.ServiceFactory{
			func(ctx app.ServiceContext) app.Service { return ecrsvc.New(ctx) },
			func(ctx app.ServiceContext) app.Service { return route53svc.New(ctx) },
			func(ctx app.ServiceContext) app.Service { return elbsvc.New(ctx) },
		}

		app, err := app.New(newConfig, factories)
		if err != nil {
			log.Fatal(err)
		}

		if err := app.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Hibiscus",
	Long:  `All software has versions. This is Hibiscus's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("🌺 Hibiscus\n\n")
		fmt.Printf("Version: %s\n", buildVersion)
		fmt.Printf("Build Date: %s\n", buildDate)
		fmt.Printf("OS/Arch: %s/%s\n", buildOS, buildArch)
		fmt.Printf("Commit: %s\n", buildCommit)
		fmt.Printf("Go Version: %s\n", buildGoVersion)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
