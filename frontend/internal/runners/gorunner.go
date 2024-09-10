package runners

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type GoRunner struct {
	conf Config
	conn *amqp091.Connection
}

func NewGoRunner(conf Config, conn *amqp091.Connection) GoRunner {
	return GoRunner{
		conf: conf,
		conn: conn,
	}
}

type goRunReq struct {
	Code string `json:"code"`
}
type goRunResp struct {
	Stdout   []byte        `json:"stdout"`
	Stderr   []byte        `json:"stderr"`
	ExitCode int           `json:"exitCode"`
	TimeTook time.Duration `json:"timeTook"`
}

func (g GoRunner) Run(ctx context.Context, code string) (*RunResult, error) {
	// TODO: Add timeout to configuration
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	resp, err := publishGetResponse[goRunResp](
		ctx,
		g.conn,
		g.conf.GoSendQ,
		g.conf.GoRespQ,
		goRunReq{Code: code},
	)

	if err != nil {
		return nil, err
	}

	return &RunResult{Sstdout: resp.Stdout, Sstderr: resp.Stderr, ExitCode: resp.ExitCode, ExecutionTime: resp.TimeTook}, nil
}
