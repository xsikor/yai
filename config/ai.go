package config

const (
	openai_key         = "OPENAI_KEY"
	openai_proxy       = "OPENAI_PROXY"
	openai_temperature = "OPENAI_TEMPERATURE"
)

type AiConfig struct {
	key         string
	proxy       string
	temperature float64
}

func (c AiConfig) GetKey() string {
	return c.key
}

func (c AiConfig) GetProxy() string {
	return c.proxy
}

func (c AiConfig) GetTemperature() float64 {
	return c.temperature
}
