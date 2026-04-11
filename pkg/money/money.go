package money

import (
	"encoding/json"
	"errors"
)

var ErrMoneyNegative = errors.New("money cannot be negative")

type Money struct {
	Cents int64
}

func New(cents int64) (Money, error) {
	if cents < 0 {
		return Money{}, ErrMoneyNegative
	}
	return Money{Cents: cents}, nil
}

func (m *Money) Add(other Money) Money {
	return Money{Cents: m.Cents + other.Cents}
}

func (m *Money) UnmarshalJSON(data []byte) error {
	var cents int64

	if err := json.Unmarshal(data, &cents); err != nil {
		return err
	}

	if cents < 0 {
		return ErrMoneyNegative
	}

	m.Cents = cents
	return nil
}

func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Cents)
}
