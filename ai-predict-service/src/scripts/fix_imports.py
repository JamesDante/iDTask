import os

def fix_imports(file_path):
    with open(file_path, "r") as f:
        lines = f.readlines()

    modified = False
    with open(file_path, "w") as f:
        for line in lines:
            if line.strip().startswith("import predict_pb2"):
                # 替换为相对导入
                line = line.replace("import predict_pb2", "from . import predict_pb2")
                modified = True
            f.write(line)

    if modified:
        print(f"✅ Fixed imports in: {file_path}")
    else:
        print(f"ℹ️ No changes needed in: {file_path}")

if __name__ == "__main__":
    target_file = "ai-predict-service/src/proto/predict_pb2_grpc.py"
    if os.path.exists(target_file):
        fix_imports(target_file)
    else:
        print(f"❌ File not found: {target_file}")