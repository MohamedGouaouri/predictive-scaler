from locust import HttpUser, task, TaskSet, constant, LoadTestShape, between
import random
class BoutiqueUserTasks(TaskSet):

    def __init__(self, parent):
        super().__init__(parent)
        self.products = [
            '0PUK6V6EV0',
            '1YMWWN1N4O',
            '2ZYFJ3GM2N',
            '66VCHSJNUP',
            '6E92ZMYYFZ',
            '9SIQT8TOJO',
            'L9ECAV7KIM',
            'LS4PSXUNUM',
            'OLJCESPC7Z']
    #
    @task(1)
    def index(self):
        self.client.get("/")
    #
    @task(2)
    def setCurrency(self):
        currencies = ['EUR', 'USD', 'JPY', 'CAD']
        self.client.post("/setCurrency",
            {'currency_code': random.choice(currencies)})
    @task(10)
    def browseProduct(self):
        self.client.get("/product/" + random.choice(self.products))

    @task(2)
    def viewCart(self):
        self.client.get("/cart")

    @task(3)
    def addToCart(self):
        product = random.choice(self.products)
        self.client.get("/product/" + product)
        self.client.post("/cart", {
            'product_id': product,
            'quantity': random.choice([1,2,3,4,5,10])})

    @task(1)
    def checkout(self):
        self.addToCart()
        self.client.post("/cart/checkout", {
            'email': 'someone@example.com',
            'street_address': '1600 Amphitheatre Parkway',
            'zip_code': '94043',
            'city': 'Mountain View',
            'state': 'CA',
            'country': 'United States',
            'credit_card_number': '4432-8015-6152-0454',
            'credit_card_expiration_month': '1',
            'credit_card_expiration_year': '2039',
            'credit_card_cvv': '672',
        })


class WebsiteUser(HttpUser):
    def on_start(self):
        return super().on_start()

    def on_stop(self):
        return super().on_stop()
    host = "http://localhost:8080"
    wait_time = constant(1)
    tasks = [BoutiqueUserTasks]
    
class StagesShape(LoadTestShape):
    """
    A simply load test shape class that has different user and spawn_rate at
    different stages.

    Keyword arguments:

        stages -- A list of dicts, each representing a stage with the following keys:
            duration -- When this many seconds pass the test is advanced to the next stage
            users -- Total user count
            spawn_rate -- Number of users to start/stop per second
            stop -- A boolean that can stop that test at a specific stage

        stop_at_end -- Can be set to stop once all stages have run.
    """

    def __init__(self):
        super().__init__()
        lines = []
        with open("random-100max.req", 'r') as f:
            lines = list(map(int, f.readlines()))
            lines = [x for i,x in enumerate(lines) if i%1==0]
            self.lines = ([1]*5+lines+[1]*5)
    
    def tick(self):
        run_time = self.get_run_time()
        for _ in range(10):#
            for i, v in enumerate(self.lines):
                if run_time < (i+1)*5:
                    tick_data = (v, 100)                
                    return tick_data



### Workload for 24 hours
# class StagesShape(LoadTestShape):
#     """
#     Load test shape that mimics real-world usage patterns over 24 hours.
#     - Morning: Gradual increase in traffic
#     - Afternoon: Peak traffic
#     - Evening: Gradual decrease in traffic
#     - Night: Low traffic
#     """
    
#     def __init__(self):
#         super().__init__()
#         # Define the 24-hour pattern (time in seconds)
#         self.stages = [
#             {"duration": 2 * 3600, "users": 100, "spawn_rate": 10},   # Early morning (low traffic)
#             {"duration": 4 * 3600, "users": 500, "spawn_rate": 20},   # Morning (gradual increase)
#             {"duration": 4 * 3600, "users": 1000, "spawn_rate": 50},  # Afternoon peak
#             {"duration": 3 * 3600, "users": 500, "spawn_rate": 30},   # Evening (gradual decrease)
#             {"duration": 3 * 3600, "users": 100, "spawn_rate": 10},   # Night (low traffic)
#             {"duration": 8 * 3600, "users": 50, "spawn_rate": 5},     # Late night (very low traffic)
#         ]

#     def tick(self):
#         run_time = self.get_run_time()

#         # Iterate through the stages and apply the correct user count and spawn rate based on the time
#         for stage in self.stages:
#             if run_time < stage["duration"]:
#                 tick_data = (stage["users"], stage["spawn_rate"])
#                 return tick_data
#             run_time -= stage["duration"]  # Move to the next stage

#         # End the test after all stages are completed
#         return None