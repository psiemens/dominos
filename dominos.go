package main

import (
	"bufio"
	"bytes"
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
	options map[string]string
}

const (
	SMALL  = "10SCREEN"
	MEDIUM = "12SCREEN"
	LARGE  = "14SCREEN"
)

const (
	CHEESE    = "C"
	PEPPERONI = "P"
	BACON     = "B"
	HAM       = "H"
)

type PriceResponse map[string]float32

func (d *DominosOrder) SetDefaults() {
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

///////////////////////////////////////////////////

type Pizza struct {
	Size     string
	Toppings Toppings
}

type Toppings []string

func (p *Pizza) ConfigurePizza(pizzas *[]Pizza) {
	fmt.Print("Choose a size: s,m,l ")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	size := strings.ToLower(text)
	switch {
	case strings.Contains(size, "s"):
		p.Size = SMALL
	case strings.Contains(size, "m"):
		p.Size = MEDIUM
	case strings.Contains(size, "l"):
		p.Size = LARGE
	}
	fmt.Print("Choose toppings, comma seperated: ham,pepperoni,cheese,bacon ")
	text, _ = reader.ReadString('\n')
	toppings := strings.ToLower(text)
	p.Toppings = []string{}
	for _, topping := range strings.Split(toppings, ",") {
		p.Toppings.AddToppings(topping)
	}
}

func (t *Toppings) AddToppings(topping string) {
	switch {
	case strings.Contains(topping, "pepporoni"):
		*t = append(*t, PEPPERONI)
	case strings.Contains(topping, "ham"):
		*t = append(*t, HAM)
	case strings.Contains(topping, "cheese"):
		*t = append(*t, CHEESE)
	case strings.Contains(topping, "bacon"):
		*t = append(*t, BACON)
	}
}

///////////////////////////////////////////////////

func (d *Dominos) ChooseProducts() {
	pizzas := &[]Pizza{}
	for {
		fmt.Print("Add a pizza?: y/n ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if !strings.Contains(text, "y") {
			break
		}
		pizza := &Pizza{}
		pizza.ConfigurePizza(pizzas)
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

func (d *DominosOrder) SetPrice(prices PriceResponse) {
	d.order["Amounts"] = prices
}

func (d *DominosOrder) GetTotal() float32 {
	return d.order["Amounts"].(PriceResponse)["Payment"]
}

func (d *DominosOrder) GetAddress() map[string]string {
	return d.order["Address"].(map[string]string)
}

func ToJSON(v interface{}) ([]byte, error) {
	JSON, err := json.Marshal(v)
	if err != nil {
		return []byte{}, err
	}
	return JSON, nil
}

func (d *DominosOrder) ToJSONString() string {
	JSON, err := ToJSON(d.order)
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

func (d *Dominos) SetAddress(street, city, region, postal, resType string) {
	d.Order.SetAddress(street, city, region, postal, resType)
}

func (d *Dominos) ToJSONString() string {
	return d.Order.ToJSONString()
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
		description := strings.Replace(stores.([]interface{})[key].(map[string]interface{})["AddressDescription"].(string), "\n", "-", 1)
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

func (d *Dominos) ValidateOrder() {
	endpoint := "https://order.dominos.ca/power/validate-order"
	bodyType := "application/json"
	bodyString := d.Order.ToJSONString()
	jsonBytes := bytes.NewBuffer([]byte(bodyString))
	resp, err := http.Post(endpoint, bodyType, jsonBytes)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	jsonResponse := make(map[string]interface{})
	err = json.Unmarshal(bodyResp, &jsonResponse)
	if err != nil {
		panic(err)
	}
	json, err := ToJSON(jsonResponse)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

//

func main() {
	order := Dominos{}
	order.SetDefaults()
	order.SetAddress("3457 West 1st Avenue", "Vancouver", "BRITISH COLUMBIA", "V6R1G6", "House")
	order.SetStores()
	order.SelectStore()
	//JSON := order.ToJSONString()
	//fmt.Println(JSON)
	order.ValidateOrder()
	order.ChooseProducts()
}
