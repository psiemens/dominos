package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Type DominosOrder is a representation of all the necessary fields to complete an order
type DominosOrder struct {
	// Representation of the order JSON object
	order map[string]interface{}
	// Pizza sizes
	sizes map[string]string
	// Pizza toppings
	options map[string]string
}

type PriceResponse map[string]float32

func (d *DominosOrder) SetDefaults() {
	d.options = map[string]string{
		"cheese":    "C",
		"pepperoni": "P",
		"bacon":     "B",
		"ham":       "H",
	}
	d.sizes = map[string]string{
		"s": "10SCREEN",
		"m": "12SCREEN",
		"l": "14SCREEN",
	}
	d.order = map[string]interface{}{
		"Address":               map[string]string{},
		"Amounts":               PriceResponse{},
		"Coupons":               []string{},
		"CustomerID":            "",
		"Email":                 "",
		"Extension":             "",
		"FirstName":             "",
		"LastName":              "",
		"LanguageCode":          "en",
		"OrderChannel":          "OLO",
		"OrderID":               "",
		"OrderMethod":           "Web",
		"OrderTaker":            nil,
		"Payments":              []string{},
		"Phone":                 "",
		"Products":              []string{},
		"ServiceMethod":         "Delivery",
		"SourceOrganizationURI": "order.dominos.ca",
		"StoreID":               "",
		"Tags":                  map[string]string{},
		"Version":               "1.0",
		"NoCombine":             true,
		"Partners":              map[string]string{},
	}
}

func (d *DominosOrder) SetAddress(street, city, region, postal, resType string) {
	d.order["Address"].(map[string]string)["Street"] = street
	d.order["Address"].(map[string]string)["City"] = city
	d.order["Address"].(map[string]string)["Region"] = region
	d.order["Address"].(map[string]string)["PostalCode"] = postal
	d.order["Address"].(map[string]string)["Type"] = resType
}

func (d *DominosOrder) SetStore(storeId string) {
	d.order["StoreID"] = storeId
}

// SetPrice is a method to set the
func (d *DominosOrder) SetPrice(prices PriceResponse) {
	d.order["Amounts"] = prices
}

func (d *DominosOrder) GetTotal() float32 {
	return d.order["Amounts"].(PriceResponse)["Payment"]
}

func (d *DominosOrder) GetAddress() map[string]string {
	return d.order["Address"].(map[string]string)
}

func (d *DominosOrder) ToJSONString() string {
	JSON, err := json.Marshal(d.order)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("{ \"Order\": %s}", string(JSON))
}

///////////////////////////////////////////////////////////////////////////

type Dominos struct {
	Order     DominosOrder
	Locations map[string]interface{}
}

func (d *Dominos) SetDefaults() {
	d.Order = DominosOrder{}
	d.Order.SetDefaults()
}

func (d *Dominos) SetStores() {
	address := d.Order.GetAddress()
	city := address["City"]
	province := address["Region"]
	postalCode := address["PostalCode"]
	street := address["Street"]
	payload := map[string]string{
		"type": url.QueryEscape("Delivery"),
		"c":    url.QueryEscape(fmt.Sprintf("%s,%s%s", city, province, postalCode)),
		"s":    url.QueryEscape(street),
	}
	endpoint := fmt.Sprintf("https://order.dominos.ca/power/store-locator?type=%s&c=%s&s=%s", payload["type"], payload["c"], payload["s"])
	resp, err := http.Get(endpoint)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	locations := make(map[string]interface{})
	err = json.Unmarshal(body, &locations)
	if err != nil {
		panic(err)
	}
	d.Locations = locations
}

func (d *Dominos) SelectStore() {
	d.Order.order["Address"] = d.Locations["Address"]
	stores := d.Locations["Stores"]
	if len(stores.([]interface{})) == 0 {
		panic("Sorry, there are no Dominos locations near you.")
	}
	fmt.Println("Select the location nearest to you")
	for key := range stores.([]interface{}) {
		location := stores.([]interface{})[key]
		description := strings.Replace(location.(map[string]interface{})["AddressDescription"].(string), "\n", "-", 1)
		fmt.Printf("(%d): %s\n", key+1, description)
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Selection: ")
	text, _ := reader.ReadString('\n')
	option, err := strconv.Atoi(strings.Replace(text, "\n", "", 1))
	if err != nil || option < 1 || option > len(stores.([]interface{}))+1 {
		panic("Invalid option")
	}
	selected := strings.Replace(stores.([]interface{})[option-1].(map[string]interface{})["AddressDescription"].(string), "\n", "-", 1)
	fmt.Printf("You selected: %d - %s\n", option, selected)
	d.Order.order["StoreID"] = stores.([]interface{})[option-1].(map[string]interface{})["StoreID"].(string)
}

//

func main() {
	order := Dominos{}
	order.SetDefaults()
	order.Order.SetAddress("3457 West 1st Avenue", "Vancouver", "BRITISH COLUMBIA", "V6R1G6", "House")
	order.SetStores()
	order.SelectStore()
	JSON := order.Order.ToJSONString()
	fmt.Println(JSON)
}
