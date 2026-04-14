package cli

import (
	"strings"
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

func TestRecurringInvoicesListJSON(t *testing.T) {
	data := fa.RecurringInvoicesResponse{RecurringInvoices: []fa.RecurringInvoice{
		{URL: "https://api.freeagent.com/v2/recurring_invoices/42", Status: "active", Currency: "GBP", TotalValue: "500.00"},
	}}
	srv := newTestServer(t, "/recurring_invoices", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "recurring-invoices", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "recurring_invoices/42") {
		t.Errorf("expected recurring invoice URL in output, got: %s", out)
	}
	if !strings.Contains(out, "500.00") {
		t.Errorf("expected total_value in output, got: %s", out)
	}
}

func TestRecurringInvoicesGetJSON(t *testing.T) {
	data := fa.RecurringInvoiceResponse{RecurringInvoice: fa.RecurringInvoice{
		URL: "https://api.freeagent.com/v2/recurring_invoices/7", Status: "active", Currency: "USD", TotalValue: "999.00",
	}}
	srv := newTestServer(t, "/recurring_invoices/7", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "recurring-invoices", "get", "7"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "999.00") {
		t.Errorf("expected total_value in output, got: %s", out)
	}
	if !strings.Contains(out, "USD") {
		t.Errorf("expected currency in output, got: %s", out)
	}
}

func TestStockItemsListJSON(t *testing.T) {
	data := fa.StockItemsResponse{StockItems: []fa.StockItem{
		{URL: "https://api.freeagent.com/v2/stock_items/5", Description: "Gadget", ItemCode: "G005", SalesPrice: "29.99"},
	}}
	srv := newTestServer(t, "/stock_items", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "stock-items", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Gadget") {
		t.Errorf("expected description in output, got: %s", out)
	}
	if !strings.Contains(out, "G005") {
		t.Errorf("expected item_code in output, got: %s", out)
	}
}

func TestStockItemsGetJSON(t *testing.T) {
	data := fa.StockItemResponse{StockItem: fa.StockItem{
		URL: "https://api.freeagent.com/v2/stock_items/3", Description: "Sprocket", ItemCode: "S003", SalesPrice: "14.50",
	}}
	srv := newTestServer(t, "/stock_items/3", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "stock-items", "get", "3"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Sprocket") {
		t.Errorf("expected description in output, got: %s", out)
	}
	if !strings.Contains(out, "14.50") {
		t.Errorf("expected sales_price in output, got: %s", out)
	}
}

func TestPriceListItemsListJSON(t *testing.T) {
	data := fa.PriceListItemsResponse{PriceListItems: []fa.PriceListItem{
		{URL: "https://api.freeagent.com/v2/price_list_items/9", Description: "Design Work", Price: "75.00"},
	}}
	srv := newTestServer(t, "/price_list_items", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "price-list-items", "list"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Design Work") {
		t.Errorf("expected description in output, got: %s", out)
	}
	if !strings.Contains(out, "75.00") {
		t.Errorf("expected price in output, got: %s", out)
	}
}

func TestPriceListItemsGetJSON(t *testing.T) {
	data := fa.PriceListItemResponse{PriceListItem: fa.PriceListItem{
		URL: "https://api.freeagent.com/v2/price_list_items/11", Description: "Dev Day Rate", Price: "800.00",
	}}
	srv := newTestServer(t, "/price_list_items/11", data)
	defer srv.Close()

	app := testApp(srv.URL + "/v2")
	out, err := runCLIWithIO(t, app, cliArgsWithConfig(t, "--json", "price-list-items", "get", "11"), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Dev Day Rate") {
		t.Errorf("expected description in output, got: %s", out)
	}
	if !strings.Contains(out, "800.00") {
		t.Errorf("expected price in output, got: %s", out)
	}
}
