import joblib
import numpy as np
import os

priority_model = joblib.load("./models/priority_rf.pkl")
time_model = joblib.load("./models/time_rf.pkl")

def predict_priority_and_time(metadata: dict) -> tuple[int, float]:

    task_type = float(metadata.get("TaskType", 0))
    urgency = float(metadata.get("Urgency", 0))
    size = float(metadata.get("PayloadSize", 0))

    X = np.array([[task_type, urgency, size]])
    
    priority = int(priority_model.predict(X)[0])
    estimated_time = round(float(time_model.predict(X)[0]), 2)
    return priority, estimated_time