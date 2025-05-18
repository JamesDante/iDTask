import pandas as pd
import numpy as np
from sklearn.ensemble import RandomForestClassifier, RandomForestRegressor
import joblib
import os

#generate random set
np.random.seed(42)
N = 200

df = pd.DataFrame({
    "task_type_id": np.random.randint(0, 5, N),
    "urgency": np.random.randint(1, 10, N),
    "payload_size": np.random.randint(100, 5000, N),
})

df["priority"] = np.where(df["urgency"] > 5, np.random.randint(6, 10, N), np.random.randint(1, 5, N))
df["estimated_time"] = 0.5 + df["payload_size"] / 2000 + (10 - df["urgency"]) * 0.1 + np.random.normal(0, 0.2, N)

df.to_csv("task_data_synthetic.csv", index=False)

X = df[["task_type_id", "urgency", "payload_size"]]
y_priority = df["priority"]
y_time = df["estimated_time"]

clf = RandomForestClassifier(n_estimators=100)
clf.fit(X, y_priority)
os.makedirs("../models", exist_ok=True)
joblib.dump(clf, "./models/priority_rf.pkl")

reg = RandomForestRegressor(n_estimators=100)
reg.fit(X, y_time)
joblib.dump(reg, "./models/time_rf.pkl")

print("✅ Mock model trained and saved.")