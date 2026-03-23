package pay

import (
	"fmt"
	"strings"
)

type StrategyContext struct {
	strategyMap map[string]Strategy
}

func NewStrategyContext(strategies ...Strategy) *StrategyContext {
	ctx := &StrategyContext{
		strategyMap: make(map[string]Strategy),
	}
	for _, strategy := range strategies {
		if strategy == nil {
			continue
		}
		ctx.Put(strategy.Channel(), strategy)
	}
	return ctx
}

func (c *StrategyContext) Put(channel string, strategy Strategy) {
	if strategy == nil {
		return
	}
	c.strategyMap[strings.ToLower(strings.TrimSpace(channel))] = strategy
}

func (c *StrategyContext) Get(channel string) (Strategy, error) {
	key := strings.ToLower(strings.TrimSpace(channel))
	if key == "" {
		return nil, fmt.Errorf("pay channel is empty")
	}
	strategy, ok := c.strategyMap[key]
	if !ok {
		return nil, fmt.Errorf("pay strategy not exist, channel=%s", channel)
	}
	return strategy, nil
}
