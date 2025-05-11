import grpc
from concurrent import futures
import time
import random

import predict_pb2
import predict_pb2_grpc
from etcd3 import Etcd3Client

from config import load_config

cfg = None

class AIPredictorServicer(predict_pb2_grpc.AIPredictorServicer):
    def __init__(self):
        cfg = load_config()
        etcd_host = cfg.get("etcd", "host", fallback="localhost")
        etcd_port = cfg.getint("etcd", "port", fallback=2379)

        self.etcd_client = Etcd3Client(host=etcd_host, port=etcd_port) 

    def Predict(self, request, context):
        return predict_pb2.PredictResponse(
            priority=random.randint(1, 10),
            estimated_time=round(random.uniform(1.0, 5.0), 2),
            recommended_worker=f"worker-{random.randint(1, 3)}"
        )

def serve():
    server_port = cfg.getint("server", "port", fallback=50051)
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    predict_pb2_grpc.add_AIPredictorServicer_to_server(AIPredictorServicer(), server)
    server.add_insecure_port('[::]:{server_port}')
    server.start()
    print("gRPC AI server running at :{server_port}")
    server.wait_for_termination()

if __name__ == "__main__":
    serve()