package freeagentapi

import (
	"encoding/json"
	"testing"
)

func TestDocumentedReferenceModels(t *testing.T) {
	var manager AccountManager
	if err := json.Unmarshal([]byte(`{"url":"https://api.freeagent.com/v2/account_managers/123","name":"Bobson Dugnutt","email":"bobson@example.com"}`), &manager); err != nil || manager.Name != "Bobson Dugnutt" {
		t.Fatalf("%+v %v", manager, err)
	}
	var stock StockItem
	if err := json.Unmarshal([]byte(`{"description":"Apple","opening_quantity":"10.0","opening_balance":"1.0","stock_on_hand":"8.0","cost_of_sale_category":"https://api.freeagent.com/v2/categories/2"}`), &stock); err != nil || stock.StockOnHand != "8.0" || stock.OpeningQuantity != "10.0" {
		t.Fatalf("%+v %v", stock, err)
	}
	var client Client
	if err := json.Unmarshal([]byte(`{"id":123,"name":"Test Company","subdomain":"testcompany"}`), &client); err != nil || client.ID != 123 || client.Subdomain != "testcompany" {
		t.Fatalf("%+v %v", client, err)
	}
	var item PriceListItem
	if err := json.Unmarshal([]byte(`{"code":"A001","item_type":"Products","quantity":"1.0","price":"10.99","vat_status":"standard"}`), &item); err != nil || item.Code != "A001" || item.VATStatus != "standard" {
		t.Fatalf("%+v %v", item, err)
	}
}
