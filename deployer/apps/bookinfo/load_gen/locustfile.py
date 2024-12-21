from locust import HttpUser, task, TaskSet, between, LoadTestShape
import random

# Define the user behavior for BookInfo
class BookInfoUserTasks(TaskSet):
    @task(4)
    def view_product_page(self):
        self.client.get("/productpage", name="Product Page")
        
    @task(4)
    def view_products_api(self):
        self.client.get("/api/v1/products", name="Products Api")

    @task(2)
    def view_products_api_0(self):
        self.client.get("/api/v1/products/0", name="Products Api details")
        
    def on_start(self):
        """Called when a simulated user starts"""
        print("User is starting interactions with BookInfo")

    def on_stop(self):
        """Called when a simulated user stops"""
        print("User has stopped interactions with BookInfo")


# Define the user class that binds to the BookInfoUserTasks
class BookInfoUser(HttpUser):
    tasks = [BookInfoUserTasks]
    wait_time = between(1, 5)


# Load shape for 24-hour realistic usage pattern
class StagesShape(LoadTestShape):
    """
    Load test shape that mimics real-world usage patterns over 24 hours.
    - Morning: Gradual increase in traffic
    - Afternoon: Peak traffic
    - Evening: Gradual decrease in traffic
    - Night: Low traffic
    """
    
    def __init__(self):
        super().__init__()
        # Define the 24-hour pattern (time in seconds)
        self.stages = [
            {"duration": 2 * 3600, "users": 10, "spawn_rate": 10},   # Early morning (low traffic)
            {"duration": 4 * 3600, "users": 50, "spawn_rate": 20},   # Morning (gradual increase)
            {"duration": 4 * 3600, "users": 100, "spawn_rate": 50},  # Afternoon peak
            {"duration": 3 * 3600, "users": 50, "spawn_rate": 30},   # Evening (gradual decrease)
            {"duration": 3 * 3600, "users": 10, "spawn_rate": 10},   # Night (low traffic)
            {"duration": 8 * 3600, "users": 5, "spawn_rate": 5},     # Late night (very low traffic)
        ]

    def tick(self):
        run_time = self.get_run_time()

        for stage in self.stages:
            if run_time < stage["duration"]:
                tick_data = (stage["users"], stage["spawn_rate"])
                return tick_data
            run_time -= stage["duration"]

        return None  # End the test after all stages are completed
