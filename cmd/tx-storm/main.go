package main

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/params"
)

const (
	// clientIdentifier to advertise over the network.
	clientIdentifier = "tx-storm"
)

var (
	// Git SHA1 commit hash of the release (set via linker flags).
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, gitDate, "the transactions generator CLI")

	flags []cli.Flag
)

// init the CLI app.
func init() {

	// Flags.
	flags = []cli.Flag{
		AccountsFlag,
		TxnsRateFlag,
		NumberFlag,
		utils.MetricsEnabledFlag,
		MetricsPrometheusEndpointFlag,
	}

	// App.
	app.Action = generatorMain
	app.Version = params.VersionWithCommit(gitCommit, gitDate)

	app.Commands = []cli.Command{}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, flags...)

	app.Before = func(ctx *cli.Context) error {
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// generatorMain is the main entry point.
func generatorMain(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 1 {
		return fmt.Errorf("url expected")
	}

	SetupPrometheus(ctx)

	url := args[0]
	num, ofTotal := getNumber(ctx)
	maxTxnsPerSec := getTxnsRate(ctx)
	accs := getAccCount(ctx)

	count := accs / ofTotal
	accMin := count * num
	accMax := accMin + count

	tt := newThreads(url, num, ofTotal, maxTxnsPerSec, accMin, accMax)
	tt.Start()

	waitForSignal()
	tt.Stop()
	return nil
}

func waitForSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
}
