import ollama
import sys


def summarize_text(text):
    response = ollama.chat(
        model="llama3.1:latest",
        messages=[
            {
                "role": "system",
                "content": "You are a helpful assistant that summarizes transcriptions into bullet points."
            },
            {
                "role": "user",
                "content": f"Summarize the entire following text into markdown formatting. Be detailed, and use bullet points: {text}",
            }
        ],
    )
    return response["message"]["content"]


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 summarize.py <text>")
        sys.exit(1)

    text = sys.argv[1]
    summary = summarize_text(text)
    print(summary)
