# Harkai local media worker

Separate private HTTP process for OpenCV image preprocessing, local ONNX explicit-content detection and Vosk Spanish speech-to-text. The Go API remains responsible for authentication, rate limits, publication and storage. No user media is sent to an external inference provider.

The image model bundled with NudeNet 3.4.2 includes AGPL-3.0 license files; this component is distributed under AGPL-3.0, conservatively honoring those files rather than its conflicting package metadata. `LICENSE` contains the terms. Corresponding worker source is served at `/source` and exposed by Go at `/v1/media-tools/source`; it includes the build recipe and checksum-verified upstream download locations. Model binaries are unchanged. ONNX Runtime, OpenCV and Vosk retain their upstream licenses; the Vosk small Spanish model is provided by https://alphacephei.com/vosk/models.

Only expose the worker on a private container network. Provide a random `MEDIA_WORKER_KEY` of at least 32 characters through the deployment environment. The Go adapter requires a successful response with the known policy before any image is stored or referenced. Missing models, invalid output, busy worker, timeout and review failures deny uploads. Older media without approval is not served or attachable.

The detector checks the full image and four overlapping crops for exposed explicit anatomy with a conservative threshold. This is not perfect detection and does not prove scene relevance. Measure false positives and misses with a reviewed validation set before changing the policy; never promise zero misses. Audio is limited to 60 seconds and the transcript must be edited/confirmed before publication. No request bodies, transcripts or evidence are logged or retained by this worker.
