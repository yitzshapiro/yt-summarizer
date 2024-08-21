import whisper
import sys

def transcribe_audio(filename):
    model = whisper.load_model("tiny")
    result = model.transcribe(filename)
    return result["text"]

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 transcribe.py <audio_file>")
        sys.exit(1)

    filename = sys.argv[1]
    transcription = transcribe_audio(filename)
    print(transcription)
