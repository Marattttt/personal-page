package runners

type Config struct {
	MQAddr string `env:"MQ_ADDR, default=localhost:5672"`
	MqUser string `env:"MQ_USER, default=guest"`
	MqPass string `env:"MQ_PASS, default=guest"`

	GoSendQ string `env:"GO_SENDQ, default=gorunner"`
	GoRespQ string `env:"GO_RESPQ, default=gorunner-response"`
}
