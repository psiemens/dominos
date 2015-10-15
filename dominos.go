package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
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
	CHEESE            = "C"
	PEPPERONI         = "P"
	BROOKLYNPEPPERONI = "Xp"
	SAUSAGE           = "S"
	BEEF              = "B"
	HAM               = "H"
	BACON             = "K"
	SALAMI            = "L"
	CHICKEN           = "D"
	PHILLYSTEAK       = "St"
	ANCHOVY           = "A"
	CHEDDARMOZZA      = "Cm"
	FETA              = "Fe"
	PROVOLONE         = "Cp"
	BANANAPEPPERS     = "Z"
	BLACKOLIVES       = "R"
	GREENOLIVES       = "V"
	GREENPEPPERS      = "G"
	MUSHROOM          = "M"
	PINEAPPLE         = "N"
	ONION             = "O"
	TOMATOES          = "T"
	JALAPENOPEPPERS   = "J"
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

type Pizza map[string]interface{}

func ConfigurePizza(p Pizza) Pizza {
	fmt.Print("Choose a size - s,m,l:\n")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	size := strings.ToLower(text)
	switch {
	case strings.Contains(size, "s"):
		p["size"] = SMALL
	case strings.Contains(size, "m"):
		p["size"] = MEDIUM
	case strings.Contains(size, "l"):
		p["size"] = LARGE
	}
	fmt.Print("Choose toppings, comma seperated - ham,pepperoni,cheese,beef:\n")
	text, _ = reader.ReadString('\n')
	toppings := strings.ToLower(text)
	pizzaToppings := []string{}
	for _, topping := range strings.Split(toppings, ",") {
		switch {
		case strings.Contains(topping, "cheese"):
			pizzaToppings = append(pizzaToppings, CHEESE)
		case strings.Contains(topping, "pepperoni"):
			pizzaToppings = append(pizzaToppings, PEPPERONI)
		case strings.Contains(topping, "brooklyn pepperoni"):
			pizzaToppings = append(pizzaToppings, BROOKLYNPEPPERONI)
		case strings.Contains(topping, "sausage"):
			pizzaToppings = append(pizzaToppings, SAUSAGE)
		case strings.Contains(topping, "beef"):
			pizzaToppings = append(pizzaToppings, BEEF)
		case strings.Contains(topping, "ham"):
			pizzaToppings = append(pizzaToppings, HAM)
		case strings.Contains(topping, "bacon"):
			pizzaToppings = append(pizzaToppings, BACON)
		case strings.Contains(topping, "salami"):
			pizzaToppings = append(pizzaToppings, SALAMI)
		case strings.Contains(topping, "chicken"):
			pizzaToppings = append(pizzaToppings, CHICKEN)
		case strings.Contains(topping, "philly steak"):
			pizzaToppings = append(pizzaToppings, PHILLYSTEAK)
		case strings.Contains(topping, "anchovy"):
			pizzaToppings = append(pizzaToppings, ANCHOVY)
		case strings.Contains(topping, "cheddar/mozza"):
			pizzaToppings = append(pizzaToppings, CHEDDARMOZZA)
		case strings.Contains(topping, "feta"):
			pizzaToppings = append(pizzaToppings, FETA)
		case strings.Contains(topping, "provolone"):
			pizzaToppings = append(pizzaToppings, PROVOLONE)
		case strings.Contains(topping, "banana peppers"):
			pizzaToppings = append(pizzaToppings, BANANAPEPPERS)
		case strings.Contains(topping, "black olives"):
			pizzaToppings = append(pizzaToppings, BLACKOLIVES)
		case strings.Contains(topping, "green olives"):
			pizzaToppings = append(pizzaToppings, GREENOLIVES)
		case strings.Contains(topping, "green peppers"):
			pizzaToppings = append(pizzaToppings, GREENPEPPERS)
		case strings.Contains(topping, "mushroom"):
			pizzaToppings = append(pizzaToppings, MUSHROOM)
		case strings.Contains(topping, "pineapple"):
			pizzaToppings = append(pizzaToppings, PINEAPPLE)
		case strings.Contains(topping, "onion"):
			pizzaToppings = append(pizzaToppings, ONION)
		case strings.Contains(topping, "tomatoes"):
			pizzaToppings = append(pizzaToppings, TOMATOES)
		case strings.Contains(topping, "jalapeno peppers"):
			pizzaToppings = append(pizzaToppings, JALAPENOPEPPERS)
		}
	}
	p["toppings"] = pizzaToppings
	return p
}

///////////////////////////////////////////////////

func (d *Dominos) ChooseProducts() {
	pizzas := []Pizza{}
	for {
		fmt.Print("Add a pizza?: y/n ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if !strings.Contains(text, "y") {
			break
		}
		pizza := Pizza{}
		pizzas = append(pizzas, ConfigurePizza(pizza))
	}
	products := []map[string]interface{}{}
	for n, pizza := range pizzas {
		options := BuildOptions(pizza["toppings"].([]string))
		product := map[string]interface{}{
			"Code":         pizza["size"],
			"Qty":          1,
			"ID":           n,
			"Instructions": "",
			"isNew":        true,
			"Options":      options,
		}
		products = append(products, product)
	}
	d.Order.order["Products"] = products
}

func BuildOptions(toppings []string) map[string]map[string]string {
	options := map[string]map[string]string{}
	for _, topping := range toppings {
		options[topping] = map[string]string{
			"1/1": "1",
		}
	}
	return options
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
	Price     map[string]interface{}
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
	locationsToStore := []interface{}{}
	for _, location := range locations["Stores"].([]interface{}) {
		if location.(map[string]interface{})["IsOnlineNow"].(bool) {
			locationsToStore = append(locationsToStore, location)
		}
	}
	d.Locations = locations
	d.Locations["Stores"] = locationsToStore
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
}

func (d *Dominos) PriceOrder() {
	endpoint := "https://order.dominos.ca/power/price-order"
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
	d.Order.order["Amounts"] = jsonResponse["Order"].(map[string]interface{})["Amounts"]
}

func (d *Dominos) GetTotal() float64 {
	return d.Order.order["Amounts"].(map[string]interface{})["Payment"].(float64)
}

func stripchars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}

func (d *Dominos) SetInformation() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("First name: ")
	text, _ := reader.ReadString('\n')
	d.Order.order["FirstName"] = strings.Replace(text, "\n", "", 1)
	fmt.Println("Last name: ")
	text, _ = reader.ReadString('\n')
	d.Order.order["LastName"] = strings.Replace(text, "\n", "", 1)
	fmt.Println("Email: ")
	text, _ = reader.ReadString('\n')
	d.Order.order["Email"] = strings.Replace(text, "\n", "", 1)
	fmt.Println("Phone: ")
	text, _ = reader.ReadString('\n')
	d.Order.order["Phone"] = stripchars(strings.Replace(text, "\n", "", 1), "-. ")
}

func (d *Dominos) ConfirmOrder() {
	fmt.Printf("Order total is %g\n", d.GetTotal())
	fmt.Println("Do you want to place the order? y/n")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	if !strings.Contains(strings.ToLower(text), "y") {
		panic("YOU DONT WANT PIZZA, NO PIZZA FOR YOU")
	}
	d.SetInformation()
	//d.PlaceOrder()
}

func (d *Dominos) PlaceOrder() {
	endpoint := "https://order.dominos.ca/power/place-order"
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
	fmt.Println(resp.Status)
	fmt.Println(resp.StatusCode)
	fmt.Println(ToJSON(jsonBytes))
}

type Envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Soap    *SoapBody
}
type SoapBody struct {
	XMLName   xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Tresponse *TokenResponse
}
type TokenResponse struct {
	XMLName xml.Name `xml:"GetTrackerDataResponse"`
	Tresult *SubTokenResponse
}
type SubTokenResponse struct {
	XMLName xml.Name `xml:"OrderStatuses"`
	Tresult *[]TokenResult
}
type TokenResult struct {
	XMLName xml.Name `xml:"OrderStatus"`
	Token   string   `xml:"OrderStatus"`
}

func (d *Dominos) Tracker() {
	//d.Order.order["Phone"].(string)
	endpoint := fmt.Sprintf("https://order.dominos.ca/orderstorage/GetTrackerData?Phone=%s", "3067156976")
	resp, err := http.Get(endpoint)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	response := new(Envelope)
	err = xml.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	fmt.Println(response.Soap.Tresponse.Tresult.Tresult)
}

func main() {
	order := Dominos{}
	// order.SetDefaults()
	// order.SetAddress("3457 West 1st Avenue", "Vancouver", "BRITISH COLUMBIA", "V6R1G6", "House")
	// order.SetStores()
	// order.SelectStore()
	// order.ValidateOrder()
	// order.ChooseProducts()
	// order.ValidateOrder()
	// order.PriceOrder()
	// order.ConfirmOrder()
	order.Tracker()
}
