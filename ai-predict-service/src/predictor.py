import joblib
import numpy as np
import os

base_dir = os.path.dirname(os.path.abspath(__file__))
model_path = os.path.join(base_dir, "../models/priority_rf.pkl")
time_model_path = os.path.join(base_dir, "../models/time_rf.pkl")

priority_model = joblib.load(model_path)
time_model = joblib.load(time_model_path)

def predict_priority_and_time(metadata: dict) -> tuple[int, float]:

    task_type = float(metadata.get("TaskType", 0))
    urgency = float(metadata.get("Urgency", 0))
    size = float(metadata.get("PayloadSize", 0))

    X = np.array([[task_type, urgency, size]])
    
    priority = int(priority_model.predict(X)[0])
    estimated_time = round(float(time_model.predict(X)[0]), 2)
    return priority, estimated_time