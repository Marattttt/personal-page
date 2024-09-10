package runners

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type RunResult struct {
	Sstdout       []byte
	Sstderr       []byte
	ExitCode      int
	ExecutionTime time.Duration
}

func publishGetResponse[R any](
	ctx context.Context,
	conn *amqp091.Connection,
	sendq string,
	recvq string,
	sendObj any,
) (*R, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("creating mq chan: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(sendq, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declaring send q: %w", err)
	}

	marshalled, err := json.Marshal(sendObj)
	if err != nil {
		return nil, fmt.Errorf("formatting send msg: %w", err)
	}

	correlationId := uuid.New()

	err = ch.Publish("", q.Name, true, false, amqp091.Publishing{
		ContentType:   "application/json",
		Body:          marshalled,
		CorrelationId: correlationId.String(),
	})

	if err != nil {
		return nil, fmt.Errorf("publishing a message: %w", err)
	}

	/*** Receive response ***/

	deliv, err := ch.Consume(recvq, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consuming: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("ctx cancelled")

		case msg := <-deliv:
			if msg.CorrelationId != correlationId.String() {
				msg.Nack(false, true)
				continue
			}

			var resp R

			if err := json.Unmarshal(msg.Body, &resp); err != nil {
				slog.Error("Could not unmarshal message from broker: %w", err)
				return nil, fmt.Errorf("unmarshalling msg %s: %w", msg.Body, err)
			}

			msg.Ack(false)

			return &resp, nil
		}
	}

}
