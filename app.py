from flask import Flask, request, jsonify
import os
import whisper
import ollama
from tqdm import tqdm
from flask_cors import CORS
import yt_dlp as youtube_dl
import logging

logging.basicConfig(level=logging.DEBUG)

app = Flask(__name__)
CORS(app)

def download_audio(url):
    ydl_opts = {
        'format': 'bestaudio/best',
        'outtmpl': 'downloaded_audio.%(ext)s',
        'postprocessors': [{
            'key': 'FFmpegExtractAudio',
            'preferredcodec': 'mp3',
            'preferredquality': '192',
        }],
        'verbose': True,
    }
    with youtube_dl.YoutubeDL(ydl_opts) as ydl:
        ydl.download([url])

def transcribe_audio(filename):
    model = whisper.load_model("tiny")
    result = model.transcribe(filename)
    return result["text"]

def summarize_text(text):
    response = ollama.chat(
        model='qwen2:1.5b',
        messages=[
            {
                'role': 'user',
                'content': f'Summarize the entire following text into markdown formatting. Be detailed: {text}',
            }
        ]
    )
    return response['message']['content']

@app.route('/process_video', methods=['POST'])
def process_video():
    data = request.json
    if 'url' not in data:
        return jsonify({'error': 'URL not provided'}), 400
    
    url = data['url']
    try:
        download_audio(url)
    except Exception as e:
        return jsonify({'error': f'Failed to download audio: {str(e)}'}), 500

    audio_file = 'downloaded_audio.mp3'
    try:
        transcription = transcribe_audio(audio_file)
    except Exception as e:
        return jsonify({'error': f'Failed to transcribe audio: {str(e)}'}), 500

    try:
        summary = summarize_text(transcription)
    except Exception as e:
        return jsonify({'error': f'Failed to summarize transcription: {str(e)}'}), 500
    
    os.remove(audio_file)

    return jsonify({'summary': summary})

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=5001)
