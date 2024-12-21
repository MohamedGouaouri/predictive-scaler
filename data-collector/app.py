import networkx as nx
import numpy as np
import torch
import pika
import json
from torch_geometric.utils.convert import from_networkx
from torch_geometric_temporal.signal import DynamicGraphTemporalSignal
import pickle
import os


# Define connection parameters to connect to RabbitMQ server
queue_url = os.getenv("QUEUE_URL", "localhost")
queue_name = os.getenv("QUEUE_NAME", "scaler")

connection_params = pika.ConnectionParameters(queue_url)
# Establish a connection with RabbitMQ server
connection = pika.BlockingConnection(connection_params)
# Create a channel
channel = connection.channel()
# Declare the queue from which to consume
queue_name = queue_name

node_features = []
edge_indices = []
edge_weights = []
max_timesteps = 200
checkpoint_interval = 10

def transform_to_nx(nx_graph: nx.DiGraph, node, parent=None):
    # Add the current node to the graph
    node_id = node["id"]
    nx_graph.add_node(node_id, 
                        workload=node.get("workload"), 
                        cpu_=node.get("cpu", 0), 
                        memory=node.get("memory", 0),
                        replicas=node.get("replicas", 0),
                        IsGateway=node.get("IsGateway", False)
                    )

    # Add an edge from the parent to this node, if there's a parent
    if parent:
        edges = parent.get("edges", [])
        for edge in edges:
            nx_graph.add_edge(parent["id"], node_id, responseTime=edge.get("responseTime", 0), requestRate=edge.get("requestRate", 0))

    # Recursively add children nodes
    children = node.get("children", [])
    for child in children:
        transform_to_nx(nx_graph, child, node)
            
    return nx_graph

def as_pyg(G):
    data = from_networkx(G)
    data.x = torch.stack([data.cpu_, data.memory]).T
    data.edge_weights = torch.stack([data.responseTime, data.requestRate]).T
    return data  # contains x, edge_index, and edge weights

def save_checkpoint(step, node_features, edge_indices, edge_weights):
    st_ms_dataset = DynamicGraphTemporalSignal(
        edge_indices=edge_indices,
        edge_weights=edge_weights,
        features=node_features,
        targets=[None] * step
    )
    filename = f'checkpoint_{step}.pkl'
    with open(filename, 'wb') as f:
        pickle.dump(st_ms_dataset, f)
    print(f'Checkpoint saved: {filename}')

def callback(ch, method, properties, body):
    print(body)
    graph = json.loads(body)
    nx_graph = nx.DiGraph()
    G = transform_to_nx(nx_graph, graph)
    data = as_pyg(G)
    x, edge_index, e = data.x, data.edge_index, data.edge_weights
    node_features.append(x)
    edge_indices.append(edge_index)
    edge_weights.append(e)
    step = len(node_features)
    print(f"Graph: {data}")
    if step % checkpoint_interval == 0:
        save_checkpoint(step, node_features, edge_indices, edge_weights)

    if step == max_timesteps:
        save_checkpoint(step, node_features, edge_indices, edge_weights)
        channel.stop_consuming()


def main():
    try:
        # Check if the queue exists
        channel.queue_declare(queue=queue_name, passive=True)
        print(f"Queue '{queue_name}' exists. Starting consumption.")
        
        # Start consuming messages
        channel.basic_consume(
            queue=queue_name,
            on_message_callback=callback,
            auto_ack=True
        )
        channel.start_consuming()
    except pika.exceptions.ChannelClosedByBroker:
        print(f"Queue '{queue_name}' does not exist. Exiting gracefully.")
        connection.close()
    except Exception as e:
        print(f"An unexpected error occurred: {e}")
        connection.close()

if __name__ == "__main__":
    main()