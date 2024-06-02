package broker

import (
	"encoding/json"
	"time"
)

type Config struct {
	InternalName   string
	Key            string
	Secret         string
	RetryTimerHTTP time.Duration
}

func (b *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		RetryTimerHTTP string `json:"RetryTimerHTTP"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert RetryTimerHTTP from string to time.Duration
	duration, err := time.ParseDuration(aux.RetryTimerHTTP)
	if err != nil {
		return err
	}
	b.RetryTimerHTTP = duration
	return nil
}
