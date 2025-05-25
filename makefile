PYTHON_VENV = ai-predict-service/venv/bin/python
PYTHON := $(shell command -v python3)

# install:
# 	$(PYTHON_VENV) -m pip install --upgrade pip
# 	$(PYTHON_VENV) -m pip install -r ai-predict-service/requirements.txt
# 	@echo "✅ Python dependencies installed"

# Set up Python virtual environment
venv:
	@echo "🐍 Setting up Python virtual environment..."
	cd ai-predict-service && \
		[ -d venv ] || $(PYTHON) -m venv venv
	$(PYTHON_VENV) -m pip install --upgrade pip
	$(PYTHON_VENV) -m pip install -r ai-predict-service/requirements.txt
	@echo "✅ Python venv created and dependencies installed"

up:
	docker compose up -d
	sleep 3
	@echo "✅ Docker containers started"

down:
	docker compose down

proto: venv
	mkdir -p ai-predict-service/src/proto
	touch ai-predict-service/src/proto/__init__.py
	protoc \
		--proto_path=proto \
		--go_out=idtask-scheduler/internal/aiclient \
		--go-grpc_out=idtask-scheduler/internal/aiclient \
		proto/predict.proto && \
		$(PYTHON_VENV) -m grpc_tools.protoc \
		-I proto \
		--python_out=ai-predict-service/src/proto \
		--grpc_python_out=ai-predict-service/src/proto \
		proto/predict.proto
		$(PYTHON_VENV) ai-predict-service/src/scripts/fix_imports.py
	@echo "✅ Proto files generated"

ai:
	PYTHONPATH=ai-predict-service/src ./ai-predict-service/venv/bin/python ai-predict-service/src/server.py

api:
	cd idtask-scheduler/api && go run .

scheduler:
	cd idtask-scheduler/scheduler && go run .

worker:
	cd idtask-scheduler/worker && go run .

client:
	cd idtask-client && npm install && npm run dev

OS := $(shell uname)

dev: venv up proto
	@echo "Starting all services..."

ifeq ($(OS), Darwin)  # macOS
	osascript -e 'tell app "Terminal" to do script "cd $(CURDIR) && make ai"'
	osascript -e 'tell app "Terminal" to do script "cd $(CURDIR) && make api"'
	osascript -e 'tell app "Terminal" to do script "cd $(CURDIR) && make scheduler"'
	osascript -e 'tell app "Terminal" to do script "cd $(CURDIR) && make worker"'
	osascript -e 'tell app "Terminal" to do script "cd $(CURDIR) && make client"'

else ifeq ($(OS), Linux)  # Linux (GNOME)
	gnome-terminal -- bash -c "make ai"
	gnome-terminal -- bash -c "make api"
	gnome-terminal -- bash -c "make scheduler"
	gnome-terminal -- bash -c "make worker"
	gnome-terminal -- bash -c "make client"

else
	@echo "❌ Unsupported OS: $(OS). Please start services manually."
endif

.PHONY: up down proto ai api scheduler worker dev install client