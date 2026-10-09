"""Local models. No user image or audio leaves this process."""
import hashlib
from pathlib import Path
import onnxruntime as ort
from vosk import Model

IMAGE_SHA = "c15d8273adad2d0a92f014cc69ab2d6c311a06777a55545f2c4eb46f51911f0f"

def load_models(directory):
    path = Path(directory) / "320n.onnx"
    if hashlib.sha256(path.read_bytes()).hexdigest() != IMAGE_SHA:
        raise RuntimeError("invalid image model")
    options = ort.SessionOptions()
    options.intra_op_num_threads = 1
    options.inter_op_num_threads = 1
    options.log_severity_level = 3
    detector = ort.InferenceSession(str(path), sess_options=options, providers=["CPUExecutionProvider"])
    return detector, Model(str(Path(directory) / "vosk-model-small-es-0.42"))
