import json
import subprocess
from vosk import KaldiRecognizer

def transcribe(model, raw):
    # Decode only local media formats; playlists/network protocols are excluded.
    command = ["ffmpeg", "-hide_banner", "-loglevel", "error", "-protocol_whitelist", "pipe", "-format_whitelist", "wav,matroska,ogg,mp3,mov,aac", "-i", "pipe:0", "-t", "61", "-vn", "-ar", "16000", "-ac", "1", "-f", "s16le", "pipe:1"]
    decoded = subprocess.run(command, input=raw, capture_output=True, timeout=15, check=False)
    pcm = decoded.stdout
    if decoded.returncode or not pcm or len(pcm) > 60*16000*2:
        raise ValueError("audio must contain at most 60 seconds")
    recognizer = KaldiRecognizer(model, 16000)
    chunks = []
    for offset in range(0, len(pcm), 8000):
        if recognizer.AcceptWaveform(pcm[offset:offset+8000]):
            chunks.append(json.loads(recognizer.Result()).get("text", ""))
    chunks.append(json.loads(recognizer.FinalResult()).get("text", ""))
    text = " ".join(chunks).strip()
    if len(text) < 5 or len(text) > 2000:
        raise ValueError("speech not understood")
    return {"text": text, "requires_confirmation": True, "source": "local_stt"}
