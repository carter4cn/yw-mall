package channel

import "fmt"

// New returns the PayChannel registered under name. Empty name means mock.
func New(name string) (PayChannel, error) {
	switch name {
	case "mock", "":
		return &MockChannel{}, nil
	case "wechat":
		return nil, fmt.Errorf("wechat channel not yet implemented (S11)")
	case "alipay":
		return nil, fmt.Errorf("alipay channel not yet implemented (S11)")
	default:
		return nil, fmt.Errorf("unknown channel: %s", name)
	}
}
