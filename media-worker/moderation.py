import cv2
import numpy as np

POLICY = "explicit-v1"
# Exposed buttocks, female breast/genitals, anus, male genitals.
EXPLICIT_ROWS = [6, 7, 8, 10, 18]

def tensor(image):
    height, width = image.shape[:2]
    side = max(height, width)
    square = cv2.copyMakeBorder(image, 0, side-height, 0, side-width, cv2.BORDER_CONSTANT, value=(0, 0, 0))
    return cv2.dnn.blobFromImage(square, 1/255, (320, 320), swapRB=True, crop=False)

def explicit(detector, image):
    output = detector.run(None, {"images": tensor(image)})[0]
    if output.shape != (1, 22, 2100) or not np.isfinite(output).all():
        raise RuntimeError("invalid detector output")
    return bool(np.max(output[0, EXPLICIT_ROWS, :]) >= 0.25)

def review(detector, raw):
    image = cv2.imdecode(np.frombuffer(raw, dtype=np.uint8), cv2.IMREAD_COLOR)
    if image is None:
        raise ValueError("invalid image")
    height, width = image.shape[:2]
    if height*width > 12_000_000 or max(height, width) > 4096 or min(height, width) < 16:
        raise ValueError("image dimensions")
    crops = [image]
    if min(height, width) >= 160:
        for y in [0, height//3]:
            for x in [0, width//3]:
                crops.append(image[y:y+2*height//3, x:x+2*width//3])
    approved = not any(explicit(detector, crop) for crop in crops)
    return {"approved": approved, "policy": POLICY, "reason": "approved" if approved else "explicit_content"}
