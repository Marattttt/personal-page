package runners

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type JsRunner struct {
	conf Config
	conn *amqp091.Connection
}

func NewJsRunner(conf Config, conn *amqp091.Connection) JsRunner {
	return JsRunner{
		conf: conf,
		conn: conn,
	}
}

type jsRunReq struct {
	Code string `json:"code"`
}

type jsRunResp struct {
	Stdout   []byte        `json:"stdout"`
	Stderr   []byte        `json:"stderr"`
	ExitCode int           `json:"exitCode"`
	TimeTook time.Duration `json:"timeTook"`
}

func (g JsRunner) Run(ctx context.Context, code string) (*RunResult, error) {
	// TODO: Add timeout to configuration
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	fmt.Println(
		g.conf.JsSendQ,
		g.conf.JsRespQ,
	)
	resp, err := publishGetResponse[jsRunResp](
		ctx,
		g.conn,
		g.conf.JsSendQ,
		g.conf.JsRespQ,
		jsRunReq{Code: code},
	)

	if err != nil {
		return nil, err
	}

	return &RunResult{Sstdout: resp.Stdout, Sstderr: resp.Stderr, ExitCode: resp.ExitCode, ExecutionTime: resp.TimeTook}, nil
}
