package main

import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/empirefox/firmata/pkg/grpci"
	"github.com/empirefox/firmata/pkg/pb"
	"github.com/empirefox/firmata/pkg/pbload"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type options struct {
	JustLoadJson    bool   `env:"PLANET_JUST_LOAD_JSON"   short:"j" long:"just-load-json"                              description:"Print json then exit"`
	Bind            string `env:"PLANET_BIND"             short:"b" long:"bind"             default:":2525"            description:"Bind address"`
	BoardsDir       string `env:"PLANET_BOARDS_DIR"       short:"s" long:"boards"           default:"/var/planet"      description:"Boards directory"`
	EtcDir          string `env:"PLANET_ETC_DIR"          short:"e" long:"etc-dir"          default:"/etc/planet"      description:"Etc directory"`
	IntegrationName string `env:"PLANET_INTEGRATION_NAME" short:"i" long:"integration-name" default:"integration.json" description:"Integration file name under etc"`
	ConfigName      string `env:"PLANET_CONFIG_NAME"      short:"c" long:"config-name"      default:"config.json"      description:"Config file name under etc"`
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	c := options{}
	_, err := flags.Parse(&c)
	if err != nil {
		// print help
		return nil
	}

	boards, err := pbload.LoadBoards(c.BoardsDir)
	if err != nil {
		return err
	}

	integration, err := pbload.LoadIntegration(filepath.Join(c.EtcDir, c.IntegrationName))
	if err != nil {
		return err
	}

	config, err := pbload.LoadConfig(filepath.Join(c.EtcDir, c.ConfigName))
	if err != nil {
		return err
	}

	// TODO add validation before here
	err = pbload.CheckError(boards, integration, config)
	if err != nil {
		return err
	}

	if c.JustLoadJson {
		// TODO pretty print
		logger.Info().Msgf("boards = %#v\n", boards)
		logger.Info().Msgf("integration = %#v\n", integration)
		logger.Info().Msgf("config = %#v\n", config)
		return nil
	}

	lis, err := net.Listen("tcp", c.Bind)
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterTransportServer(grpcServer, grpci.NewServer(ctx,
		&logger, pbload.LoadApiVersion(), boards, integration, config,
	))
	return grpcServer.Serve(lis)
}

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}
