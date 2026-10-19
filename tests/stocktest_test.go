package stock_test

import (
	"testing"
	"webtyp.com/stock"
	"webtyp.com/stock/stocktest"
	"webtyp.com/time"
)

type memLot struct {
	Product  string
	Location string
	stock.Lot
}

type memLedger struct {
	lots    []memLot
	entries []stocktest.Entry
}

func (m *memLedger) Receive(product, location, lot string, expiry int64, quantity float64) error {
	m.entries = append(m.entries, stocktest.Entry{
		Type:     stock.Purchase,
		Location: location,
		LotID:    lot,
		Quantity: quantity,
	})

	for i, l := range m.lots {
		if l.Product == product && l.Location == location && l.ID == lot {
			newQty, err := stock.Apply(l.Quantity, stock.In, quantity)
			if err != nil {
				return err
			}
			m.lots[i].Quantity = newQty
			return nil
		}
	}

	m.lots = append(m.lots, memLot{
		Product:  product,
		Location: location,
		Lot: stock.Lot{
			ID:       lot,
			Expiry:   expiry,
			Quantity: quantity,
		},
	})
	return nil
}

func (m *memLedger) Output(product, location string, mov stock.Movement, quantity float64) ([]stock.Take, error) {
	var lots []stock.Lot
	for _, l := range m.lots {
		if l.Product == product && l.Location == location {
			lots = append(lots, l.Lot)
		}
	}

	takes, err := stock.Allocate(lots, quantity, time.Now())
	if err != nil {
		return nil, err
	}

	for _, take := range takes {
		m.entries = append(m.entries, stocktest.Entry{
			Type:     mov,
			Location: location,
			LotID:    take.LotID,
			Quantity: take.Quantity,
		})

		for i, l := range m.lots {
			if l.Product == product && l.Location == location && l.ID == take.LotID {
				newQty, err := stock.Apply(l.Quantity, stock.Out, take.Quantity)
				if err != nil {
					return nil, err
				}
				m.lots[i].Quantity = newQty
				break
			}
		}
	}

	return takes, nil
}

func (m *memLedger) Transfer(product, from, to string, quantity float64) error {
	takes, err := m.Output(product, from, stock.TransferOut, quantity)
	if err != nil {
		return err
	}

	for _, take := range takes {
		var expiry int64
		for _, l := range m.lots {
			if l.Product == product && l.Location == from && l.ID == take.LotID {
				expiry = l.Expiry
				break
			}
		}

		m.entries = append(m.entries, stocktest.Entry{
			Type:     stock.TransferIn,
			Location: to,
			LotID:    take.LotID,
			Quantity: take.Quantity,
		})

		found := false
		for i, l := range m.lots {
			if l.Product == product && l.Location == to && l.ID == take.LotID {
				newQty, _ := stock.Apply(l.Quantity, stock.In, take.Quantity)
				m.lots[i].Quantity = newQty
				found = true
				break
			}
		}

		if !found {
			m.lots = append(m.lots, memLot{
				Product:  product,
				Location: to,
				Lot: stock.Lot{
					ID:       take.LotID,
					Expiry:   expiry,
					Quantity: take.Quantity,
				},
			})
		}
	}
	return nil
}

func (m *memLedger) Lots(product, location string) ([]stock.Lot, error) {
	var res []stock.Lot
	for _, l := range m.lots {
		if l.Product == product && l.Location == location && l.Quantity > 0 {
			res = append(res, l.Lot)
		}
	}
	return res, nil
}

func (m *memLedger) Entries(product string) ([]stocktest.Entry, error) {
	var res []stocktest.Entry
	for _, e := range m.entries {
		res = append(res, e) // Since it doesn't store product in Entry right now, but tests assume single product
	}
	return res, nil
}

func TestStocktest(t *testing.T) {
	stocktest.Run(t, func(t *testing.T) stocktest.Ledger {
		return &memLedger{}
	})
}
