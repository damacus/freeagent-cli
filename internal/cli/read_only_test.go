package cli

import (
	"testing"

	fa "github.com/damacus/freeagent-cli/internal/freeagentapi"
)

func TestRecurringInvoicesList(t *testing.T) {
	data := fa.RecurringInvoicesResponse{RecurringInvoices: []fa.RecurringInvoice{
		{URL: "https://api.freeagent.com/v2/recurring_invoices/1", Status: "active", Currency: "GBP"},
	}}
	srv := newTestServer(t, "/recurring_invoices", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "recurring-invoices", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStockItemsList(t *testing.T) {
	data := fa.StockItemsResponse{StockItems: []fa.StockItem{
		{URL: "https://api.freeagent.com/v2/stock_items/1", Description: "Widget", ItemCode: "W001", SalesPrice: "9.99"},
	}}
	srv := newTestServer(t, "/stock_items", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "stock-items", "list"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPriceListItemsList(t *testing.T) {
	data := fa.PriceListItemsResponse{PriceListItems: []fa.PriceListItem{
		{URL: "https://api.freeagent.com/v2/price_list_items/1", Description: "Consulting", Price: "150.00"},
	}}
	srv := newTestServer(t, "/price_list_items", data)
	defer srv.Close()
	err := testApp(srv.URL).Run([]string{"fa", "--json", "price-list-items", "list"})
	if err != nil {
		t.Fatal(err)
	}
}
