FROM python:3.12-slim
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg libgomp1 && rm -rf /var/lib/apt/lists/*
WORKDIR /worker
COPY media-worker/requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY media-worker/ ./
RUN python download_models.py && useradd --no-create-home worker
USER worker
ENV OMP_NUM_THREADS=1 MODEL_DIRECTORY=/models
EXPOSE 8090
HEALTHCHECK --interval=30s --timeout=5s CMD python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8090/health',timeout=3)"
CMD ["python", "server.py"]
