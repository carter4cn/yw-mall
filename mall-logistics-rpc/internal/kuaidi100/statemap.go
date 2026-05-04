package kuaidi100

const (
	StateInternalCreated    int32 = 0
	StateInternalCollected  int32 = 1
	StateInternalInTransit  int32 = 2
	StateInternalDelivering int32 = 3
	StateInternalDelivered  int32 = 4
	StateInternalException  int32 = 5
	StateInternalReturned   int32 = 6
	StateKuaidi100Synthetic int32 = 255
)

// MapState maps a kuaidi100 state code to internal status enum.
func MapState(k int32) int32 {
	switch k {
	case 1:
		return StateInternalCollected
	case 0:
		return StateInternalInTransit
	case 5:
		return StateInternalDelivering
	case 3:
		return StateInternalDelivered
	case 2:
		return StateInternalException
	case 4, 6, 14:
		return StateInternalReturned
	case StateKuaidi100Synthetic:
		return StateInternalException
	default:
		return StateInternalInTransit
	}
}
