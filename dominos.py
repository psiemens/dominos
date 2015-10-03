import click
import requests
import json

@click.command()
@click.option('--address', prompt=True)
@click.option('--city', prompt=True)
@click.option('--province', prompt=True)
@click.option('--postal_code', prompt=True)

class Dominos(object):

	def __init__(self, address, city, province, postal_code):

		self.order = DominosOrder()

		try:	
			self.place_order(address, city, province, postal_code)
		except PizzaProblem as e:
			click.echo(str(e))
			return

	def place_order(self, address, city, province, postal_code):

		self.select_store(address, city, province, postal_code)

		self.validate_order()

		self.choose_products()

		#self.order.set_products([{"Code":"14SCEXTRAV","Qty":1,"ID":1,"Instructions":"","isNew":True,"Options":{"X":{"1/1":"1"},"C":{"1/1":"1.5"},"P":{"1/1":"1"},"S":{"1/1":"1"},"B":{"1/1":"1"},"H":{"1/1":"1"},"R":{"1/1":"1"},"G":{"1/1":"1"},"M":{"1/1":"1"},"O":{"1/1":"1"}}}])

		self.price_order()

		if click.confirm("Your total is $%s. Place order?" % self.order.get_payment(), abort=True):
			click.echo('Your order has been placed!')


	def select_store(self, address, city, province, postal_code):

		payload = {
			'type': 'Delivery',
			'c': "%s, %s %s" % (city, province, postal_code), 
			's': address
		}

		response = requests.get('https://order.dominos.ca/power/store-locator?type=Delivery', params=payload).json()

		self.order.set_address(response['Address'])

		stores = response["Stores"]

		if len(stores) == 0:
			raise PizzaProblem("Sorry, there are no Domino's locations near you.")

		for n, store in enumerate(stores):
			click.echo("%d) %s" % (n+1, store["AddressDescription"].split('\n')[0]))

		store_index = click.prompt('Choose a location', type=int)

		if store_index < 1 or store_index > len(stores):
			raise PizzaProblem('Not a valid store option.')

		self.order.set_store(stores[store_index-1])

	def choose_products(self):

		def choose_product(pizzas=[]):
			pizza = {}
			pizza['size'] = click.prompt('Choose a size (s, m, l)', type=str)

			def choose_toppings():
				choices = click.prompt('Choose toppings (comma-separated list). Type "options" for selection', type=str)
				if choices == 'options':
					click.echo("cheese, pepperoni, ham, pineapple")
					choose_toppings()
				else:
					return choices

			pizza['toppings'] = choose_toppings()

			pizzas.append(pizza)

			if click.confirm("Do you want to add another pizza?"):
				return choose_product(pizzas)
			else:
				return pizzas

		self.order.set_products(choose_product())

	def validate_order(self):
		response = requests.post("https://order.dominos.ca/power/validate-order", json=self.order.json())

	def price_order(self):
		response = requests.post("https://order.dominos.ca/power/price-order", json=self.order.json())
		self.order.set_price(response.json())

class DominosOrder(object):

	def __init__(self):

		self.order = {
			"Address": None,
			"Coupons":[],
			"CustomerID":"",
			"Email":"",
			"Extension":"",
			"FirstName":"",
			"LastName":"",
			"LanguageCode":"en",
			"OrderChannel":"OLO",
			"OrderID":"",
			"OrderMethod":"Web",
			"OrderTaker": None,
			"Payments":[],
			"Phone":"",
			"Products":[],
			"ServiceMethod":"Delivery",
			"SourceOrganizationURI":"order.dominos.ca",
			"StoreID": None,
			"Tags":{},
			"Version":"1.0",
			"NoCombine": True,
			"Partners":{}
		}

	def set_address(self, address):
		self.order["Address"] = address

	def set_store(self, store):
		self.order["StoreID"] = store["StoreID"]

	def set_products(self, pizzas):
		for pizza in pizzas:
			click.echo(pizza)
		#self.order["Products"] = products

	def set_price(self, response):
		self.order["Amounts"] = response["Order"]["Amounts"]

	def get_payment(self):
		return self.order["Amounts"]["Payment"]

	def json(self):
		return { "Order": self.order }

class PizzaProblem(Exception):
    pass

if __name__ == '__main__':
    d = Dominos()
    d.validate_address()