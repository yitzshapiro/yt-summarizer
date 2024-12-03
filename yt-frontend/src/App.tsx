import { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { Button } from '@nextui-org/button';
import { Input } from "@nextui-org/input";
import { Card, CardBody } from '@nextui-org/card';
import './styles/globals.css';
import remarkGfm from 'remark-gfm';

function App() {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [accumulatedMarkdown, setAccumulatedMarkdown] = useState('');
  const [error, setError] = useState('');
  const [status, setStatus] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setAccumulatedMarkdown('');
    setError('');
    setStatus('');

    if (!isValidYouTubeUrl(url)) {
      setError('Please enter a valid YouTube URL');
      setLoading(false);
      return;
    }

    const eventSource = new EventSource(`http://localhost:5001/process_video?url=${encodeURIComponent(url)}`);
    let tempMarkdown = '';

    eventSource.addEventListener('status', (event) => {
      setStatus(event.data);
      if (event.data === 'Completed') {
        setLoading(false);
        eventSource.close();
      }
    });

    eventSource.addEventListener('result', (event) => {
      console.log('Received chunk:', event.data);
      tempMarkdown += event.data; // Accumulate data without adding new lines
      setAccumulatedMarkdown(tempMarkdown);
      console.log('Updated markdown:', tempMarkdown);
    });

    eventSource.addEventListener('error', (event: Event) => {
      const errorEvent = event as MessageEvent;
      setError(errorEvent.data || 'An error occurred while processing the video.');
      setLoading(false);
      eventSource.close();
    });

    eventSource.onerror = (event) => {
      console.error('EventSource failed:', event);
      setError('An error occurred while processing the video. Please try again.');
      setLoading(false);
      eventSource.close();
    };
  };

  const isValidYouTubeUrl = (url: string): boolean => {
    const youtubeRegex = /^(https?:\/\/)?(www\.)?(youtube\.com|youtu\.?be)\/.+$/;
    return youtubeRegex.test(url);
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <div className="p-8 rounded-lg shadow-lg w-full max-w-xl" style={{ marginTop: '50px'}}>
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

        {status && (
          <div className="text-blue-500 mb-4">
            {status}
          </div>
        )}

        {!loading && accumulatedMarkdown && (
          <Card className="mt-4 p-2">
            <CardBody>
              <h3>Raw Markdown:</h3>
              <pre>{accumulatedMarkdown}</pre>
              <h3>Rendered Markdown:</h3>
              <ReactMarkdown 
                remarkPlugins={[remarkGfm]}
                components={{
                  p: ({node, ...props}) => <p className="mb-4" {...props} />,
                }}
              >
                {accumulatedMarkdown}
              </ReactMarkdown>
            </CardBody>
          </Card>
        )}
      </div>
    </div>
  );
}

export default App;