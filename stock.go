package stock

// Error is a failure of a stock rule. Constants, compared by value.
type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrQuantity        Error = "stock: quantity must be greater than zero"
	ErrInsufficient    Error = "stock: not enough stock"
	ErrUnknownMovement Error = "stock: unknown movement"
)

// Direction says whether a movement adds to or takes from the stock.
type Direction int8

const (
	In  Direction = 1
	Out Direction = -1
)

// Movement is the reason stock moved. The direction is fixed by the movement.
type Movement string

const (
	Purchase    Movement = "PURCHASE"     // In
	AdjustIn    Movement = "ADJUST_IN"    // In
	TransferIn  Movement = "TRANSFER_IN"  // In
	AdjustOut   Movement = "ADJUST_OUT"   // Out
	Sale        Movement = "SALE"         // Out
	InternalUse Movement = "INTERNAL_USE" // Out
	TransferOut Movement = "TRANSFER_OUT" // Out
)

// Direction returns the direction of the movement; ok is false for a movement this package does not know.
func (m Movement) Direction() (d Direction, ok bool) {
	switch m {
	case Purchase, AdjustIn, TransferIn:
		return In, true
	case AdjustOut, Sale, InternalUse, TransferOut:
		return Out, true
	default:
		return 0, false
	}
}
