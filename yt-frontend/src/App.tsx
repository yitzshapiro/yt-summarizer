import { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { Button } from '@nextui-org/button';
import { Input } from "@nextui-org/input";
import { Card, CardBody } from '@nextui-org/card';
import './styles/globals.css';

function App() {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [markdown, setMarkdown] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMarkdown('');
    setError('');
    
    try {
      const response = await fetch('http://localhost:5001/process_video', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ url }),
      });

      if (!response.ok) {
        throw new Error('Invalid URL or server error');
      }

      const data = await response.json();
      setMarkdown(data.summary);
    } catch (error) {
      console.error('Error:', error);
      setError('An error occurred. Please ensure the URL is valid and try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <div className="p-8rounded-lg shadow-lg w-full max-w-xl" style={{ marginTop: '50px'}}>
        <h1 className="text-2xl font-bold mb-4">YouTube Video Summarizer</h1>
        <form onSubmit={handleSubmit} style={{ marginTop: '10px'}}>
          <div className="mb-4">
            <Input
              type="text"
              className="rounded w-full"
              placeholder="Enter YouTube URL"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              required
              aria-label="YouTube URL"
            />
          </div>
          <div className="flex justify-end">
          <Button
            type="submit"
            variant="solid"
            className="mb-4"
            isLoading={loading}
            disabled={loading}
            aria-busy={loading}
            style={{ marginTop: '10px' }}
          >
            Summarize
          </Button>
          </div>
        </form>

        {error && (
          <div className="text-red-500 mb-4">
            {error}
          </div>
        )}

        {!loading && markdown && (
          <Card className="mt-4 p-2">
            <CardBody>
              <ReactMarkdown>{markdown}</ReactMarkdown>
            </CardBody>
          </Card>
        )}
      </div>
    </div>
  );
}

export default App;
