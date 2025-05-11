# AI Predict Service

This is a lightweight Python service that exposes an HTTP `/predict` endpoint for task scheduling predictions.

## 🧠 Features
- Flask-based API
- Returns mocked prediction results: priority, estimated time, and recommended worker
- Designed to integrate with a Go-based scheduler

---

## 🚀 Quick Start (Recommended with Virtual Environment)

```bash
# 1. Create and activate virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# 2. Install dependencies
pip install -r requirements.txt

# 3. Run the service
python app.py
```

The service will be running at: `http://localhost:5000/predict`

---

## 🐳 Docker (Alternative)

```bash
# Build and run the container
docker build -t ai-predict-service .
docker run -p 5000:5000 ai-predict-service
```

Or use `docker-compose` if integrated into your main project.

---

## 🧪 Example Request

```bash
curl -X POST http://localhost:5000/predict \
  -H "Content-Type: application/json" \
  -d '{"task_id": "123", "metadata": {"type": "image-classify"}}'
```

## 📁 Project Structure

```
ai-predict-service/
├── app.py
├── requirements.txt
├── Dockerfile
├── venv/               # optional, virtual environment (not committed)
└── .vscode/settings.json  # optional, for VS Code interpreter config
```

---

## ✅ Tips
- Use a virtual environment to avoid `externally-managed-environment` errors.
- Activate your venv before running or installing anything.
- You can test the service locally or via Docker.

---

## 📌 License
MIT
