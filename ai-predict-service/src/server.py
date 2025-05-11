import grpc
from concurrent import futures
import time
import random

import predict_pb2
import predict_pb2_grpc
from etcd3 import Client

from config import load_config

cfg = load_config()

class AIPredictorServicer(predict_pb2_grpc.AIPredictorServicer):
    def __init__(self):
        #cfg = load_config()
        etcd_host = cfg["etcd"]["host"]
        etcd_port = cfg["etcd"]["port"]

        print(f"Connecting to etcd at {etcd_host}:{etcd_port}")

        self.etcd_client = Client(host=etcd_host, port=etcd_port) 

    def Predict(self, request, context):
        if not request.task_id:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("Task ID is required")
            return predict_pb2.PredictResponse()

        try:
            workers = self.get_workers()

            if not workers:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details("No workers available")
                return predict_pb2.PredictResponse()

            recommended_worker = random.choice(list(workers.values()))

            print(f"Recommended worker {recommended_worker}")

            priority = random.randint(1, 10)
            estimated_time = round(random.uniform(1.0, 5.0), 2)

            return predict_pb2.PredictResponse(
                priority=priority,
                estimated_time=estimated_time,
                recommended_worker=recommended_worker
            )
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Prediction failed: {str(e)}")
            return predict_pb2.PredictResponse()

    def get_workers(self, prefix='/workers/'):
        end = prefix[:-1] + chr(ord(prefix[-1]) + 1)
        response = self.etcd_client.range(prefix, end)
        return {
            kv.key.decode(): kv.value.decode()
            for kv in response.kvs
        }

    def get_workers_from_etcd(self):
        workers = []
        try:
            print(self.etcd_client.range('/workers/'))
            for value, metadata in self.etcd_client.range('/workers/'):
                worker_info = json.loads(value.decode('utf-8'))  
                workers.append(worker_info) 
        except Exception as e:
            print(f"Failed to fetch workers from etcd: {e}")
        return workers

def serve():
    global cfg 
    cfg = load_config()

    server_port = cfg["server"]["port"]

    print(f"Try running AI server at :{server_port}")

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    predict_pb2_grpc.add_AIPredictorServicer_to_server(AIPredictorServicer(), server)
    server.add_insecure_port(f'[::]:{server_port}')
    server.start()
    print(f"gRPC AI server running at :{server_port}")
    server.wait_for_termination()

if __name__ == "__main__":
    serve()