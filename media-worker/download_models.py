import hashlib
import urllib.request
import zipfile
from pathlib import Path

def download(url, sha, target):
    urllib.request.urlretrieve(url, target)
    if hashlib.sha256(target.read_bytes()).hexdigest() != sha:
        raise RuntimeError("model download checksum failed")
    return target

def install():
    directory = Path("/models")
    directory.mkdir(exist_ok=True)
    wheel = download("https://files.pythonhosted.org/packages/1c/ee/1aa02d44ba958cc77e16ff1e41a0aac5e721037db7bf62b9c9d124917f87/nudenet-3.4.2-py3-none-any.whl", "5937dbd84e5d8e5de038f08ffea5a1bb50a08475776bf2b4795914ce0eaf0331", directory / "image-model.whl")
    with zipfile.ZipFile(wheel) as archive:
        directory.joinpath("320n.onnx").write_bytes(archive.read("nudenet/320n.onnx"))
    wheel.unlink()
    speech = download("https://alphacephei.com/vosk/models/vosk-model-small-es-0.42.zip", "09b239888f633ef2f0b4e09736e3d9936acfd810bc65d53fad45261762c6511f", directory / "speech.zip")
    with zipfile.ZipFile(speech) as archive:
        archive.extractall(directory)
    speech.unlink()
    return directory

if __name__ == "__main__":
    install()
