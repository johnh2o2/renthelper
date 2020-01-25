package avalon

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"math"
)

type BalanceSheet struct {
	balance float64
}

type Tenant struct {
	Name string
	Rent float64
}

type Tenants []*Tenant

func (t Tenants) totalRent() float64 {
	sum := 0.0
	for _, ten := range t {
		sum += ten.Rent
	}
	return sum
}

// Setup. Mostly a hack for passing in RENTERS and RENT_AMOUNTS instead of a config file.
func GetTenants(renters, rentAmounts string, amenAmounts string) (Tenants, error) {
	out := Tenants{}
	rs := strings.Split(renters, ",")
	amts := strings.Split(rentAmounts, ",")
	amen := strings.Split(amenAmounts, ",")
	if len(rs) != len(amts) {
		return nil, fmt.Errorf("rent and amounts must be same size: %v vs %v", rs, amts)
	}
	if len(amen) != len(amts) {
		return nil, fmt.Errorf("amenities amounts and rent amounts must be same size: %v vs %v", amen, amts)
	}
	amounts := make([]float64, len(amts))
	for i, amt := range amts {
		amount, err := strconv.ParseFloat(amt, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %v: %v", amt, err)
		}
		amounts[i] = amount
		// Add individual amenities amount to each renter's rent
		amenAmount, err := strconv.ParseFloat(amen[i], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %v: %v", amenAmount, err)
		}
		amounts[i] += amenAmount
	}
	for i, renter := range rs {
		out = append(out, &Tenant{
			Name: renter,
			Rent: amounts[i],
		})
	}
	return out, nil
}

func (b *BalanceSheet) SplitBalance(ts Tenants) (string, error) {
	if b.balance == float64(0) {
		return "Nothing is due right now!", nil
	}
	if b.balance < ts.totalRent() {
		return fmt.Sprintf("An unusual balance. I don't know what to do with: %v", b.balance), nil
	}
	remaining := b.balance - ts.totalRent()
	evenSplitLeftover := remaining / float64(len(ts))
	for _, tenant := range ts {
		tenant.Rent += evenSplitLeftover
		tenant.Rent = math.Round(tenant.Rent*100)/100
	}
	t, err := template.New("sheet").Parse(balanceTemplate)
	var buf bytes.Buffer
	err = t.Execute(&buf, ts)
	if err != nil {
		return "", fmt.Errorf("could not execute template: %v", err)
	}
	return buf.String(), nil
}

var balanceTemplate = "```" + `{{ range $element := .}}{{.Name}}: {{.Rent}}
{{end}}` + "```"

func balanceFromHTML(balance []byte) (*BalanceSheet, error) {
	balanceRegex := regexp.MustCompile(`\$([\d.,]+)`)
	b := balanceRegex.FindSubmatch(balance)
	if b == nil {
		return nil, fmt.Errorf("Unknown balance")
	}
	// Remove commas
	cleanBytes := bytes.Replace(b[1], []byte(","), []byte(""), -1)
	val, err := strconv.ParseFloat(string(cleanBytes), 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse balance (%v): %v", b[1], err)
	}
	return &BalanceSheet{
		balance: val,
	}, nil
}
