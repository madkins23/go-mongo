package test

import (
	"fmt"
	"time"

	"github.com/madkins23/go-type/reg"
)

var _ Investment = &Stock{}
var _ Borrower = &Federal{}
var _ Borrower = &State{}

// Register adds the 'test' alias and registers several structs.
// Uses the github.com/madkins23/go-type library to register structs by name.
func Register() error {
	if err := reg.AddAlias("test", &Stock{}); err != nil {
		return fmt.Errorf("adding 'test' alias: %w", err)
	}
	if err := reg.Register(&Stock{}); err != nil {
		return fmt.Errorf("registering Stock struct: %w", err)
	}
	if err := reg.Register(&Federal{}); err != nil {
		return fmt.Errorf("registering Federal struct: %w", err)
	}
	if err := reg.Register(&State{}); err != nil {
		return fmt.Errorf("registering State struct: %w", err)
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////

type Investment interface {
	Name() string
	Key() string
	Value() float32
}

//========================================================================

const MarketNASDAQ = "NASDAQ"
const MarketNYSE = "NYSE"

type Stock struct {
	Market string
	Named  string
	Symbol string
	Shares float32
	Price  float32
}

func (s *Stock) Name() string {
	return s.Named
}

func (s *Stock) Key() string {
	return s.Symbol
}

func (s *Stock) Value() float32 {
	return s.Shares * s.Price
}

const StockCostcoName = "Costco"
const StockCostcoSymbol = "COST"
const StockCostcoShares float32 = 12.43
const StockCostcoPrice float32 = 512.10

func MakeCostco() *Stock {
	return &Stock{
		Market: MarketNASDAQ,
		Named:  StockCostcoName,
		Symbol: StockCostcoSymbol,
		Shares: StockCostcoShares,
		Price:  StockCostcoPrice,
	}
}

const StockWalmartName = "Walmart"
const StockWalmartSymbol = "WMT"
const StockWalmartShares float32 = 58.91
const StockWalmartPrice float32 = 122.26

func MakeWalmart() *Stock {
	return &Stock{
		Market: MarketNYSE,
		Named:  StockWalmartName,
		Symbol: StockWalmartSymbol,
		Shares: StockWalmartShares,
		Price:  StockWalmartPrice,
	}
}

//========================================================================

type BondData struct {
	Named    string
	Duration time.Duration
	Interest float32
	Price    float32
	Units    uint32
}

func (b *BondData) Key() string {
	return ""
}

func (b *BondData) Name() string {
	return b.Named
}

func (b *BondData) Value() float32 {
	return float32(b.Units) * b.Price
}

//////////////////////////////////////////////////////////////////////////

// Borrower is broken out to test nesting of interface objects.
// Borrower is nested within Bond.
type Borrower interface {
	Name() string
}

//========================================================================

const BondTBillName = "T-Bill"
const BondTBillDuration = 365 * 24 * time.Hour
const BondTBillInterest = 1.25
const BondTBillPrice = 0.95
const BondTBillUnits = 5000

func TBillData() BondData {
	return BondData{
		Named:    BondTBillName,
		Duration: BondTBillDuration,
		Interest: BondTBillInterest,
		Price:    BondTBillPrice,
		Units:    BondTBillUnits,
	}
}

//------------------------------------------------------------------------

type Federal struct {
}

func TBillSource() Borrower {
	return &Federal{}
}

func (f *Federal) Name() string {
	return "Treasury"
}

//========================================================================

const BondStateName = "Roads"
const BondStateDuration = 10 * 365 * 24 * time.Hour
const BondStateInterest = 2.75
const BondStatePrice = 0.98
const BondStateUnits = 1000

func StateBondData() BondData {
	return BondData{
		Named:    BondStateName,
		Duration: BondStateDuration,
		Interest: BondStateInterest,
		Price:    BondStatePrice,
		Units:    BondStateUnits,
	}
}

//------------------------------------------------------------------------

func StateBondSource() Borrower {
	return &State{State: "Confusion"}
}

type State struct {
	State string
}

func (s *State) Name() string {
	return s.State
}
