import yaml
import os

def load_config(path="config.yaml"):
    root_path = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    config_path = os.path.join(root_path, "config.yaml")

    with open(config_path, "r") as f:
        return yaml.safe_load(f)

config = load_config()