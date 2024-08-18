import React, { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import './App.css';

function App() {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [markdown, setMarkdown] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setMarkdown('');
    
    try {
      const response = await fetch('http://localhost:5001/process_video', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ url }),
      });

      const data = await response.json();
      setMarkdown(data.summary);
    } catch (error) {
      console.error('Error:', error);
      setMarkdown('An error occurred. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-6">
      <div className="bg-white p-8 rounded-lg shadow-lg w-full max-w-xl">
        <h1 className="text-2xl font-bold mb-6">YouTube Video Summarizer</h1>
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            className="border rounded p-2 w-full mb-4"
            placeholder="Enter YouTube URL"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            required
          />
          <button
            type="submit"
            className="bg-blue-500 text-white p-2 rounded w-full"
            disabled={loading}
          >
            Summarize Video
          </button>
        </form>

        {loading && (
          <div className="mt-4">
            <p className="text-center text-gray-600">Processing...</p>
            <div className="w-full bg-gray-200 rounded-full h-2.5 mt-2">
              <div className="bg-blue-500 h-2.5 rounded-full progress-bar" />
            </div>
          </div>
        )}

        {!loading && markdown && (
          <div className="mt-6 prose">
            <ReactMarkdown>{markdown}</ReactMarkdown>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;