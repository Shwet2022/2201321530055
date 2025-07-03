import React, { useState } from 'react';
import './App.css';

interface CreateURLResponse {
  shortLink: string;
  expiry: string;
}

interface URLStatsResponse {
  originalUrl: string;
  createdAt: string;
  expiresAt: string;
  hits: number;
}

interface ApiError {
  error: string;
}

const API_BASE_URL = 'http://localhost:8080';

// --- Main App Component ---
const App: React.FC = () => {
  const [view, setView] = useState<'shortener' | 'stats'>('shortener');

  return (
    <div className="container">
      <header className="header">
        <h1>URL Shortener</h1>
        <nav>
          <button onClick={() => setView('shortener')} className={view === 'shortener' ? 'active' : ''}>
            Shorten URL
          </button>
          <button onClick={() => setView('stats')} className={view === 'stats' ? 'active' : ''}>
            View Stats
          </button>
        </nav>
      </header>
      <main className="main-content">
        {view === 'shortener' ? <URLShortenerForm /> : <URLStatsPage />}
      </main>
      <footer className="footer">
        <p>Affordmed Full Stack Assessment</p>
      </footer>
    </div>
  );
};

const URLShortenerForm: React.FC = () => {
  const [longUrl, setLongUrl] = useState('');
  const [customShortcode, setCustomShortcode] = useState('');
  const [validity, setValidity] = useState(30);
  const [result, setResult] = useState<CreateURLResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResult(null);

    if (!longUrl) {
      setError('Please enter a URL to shorten.');
      setLoading(false);
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/shorturls`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          url: longUrl,
          shortcode: customShortcode || undefined,
          validity: validity,
        }),
      });

      const data: CreateURLResponse | ApiError = await response.json();

      if (!response.ok) {
        throw new Error((data as ApiError).error || 'An unknown error occurred.');
      }

      setResult(data as CreateURLResponse);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card">
      <h2>Create a Short URL</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="longUrl">Original URL</label>
          <input
            type="url"
            id="longUrl"
            value={longUrl}
            onChange={(e) => setLongUrl(e.target.value)}
            placeholder="https://example.com/very/long/url"
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="customShortcode">Custom Shortcode (Optional)</label>
          <input
            type="text"
            id="customShortcode"
            value={customShortcode}
            onChange={(e) => setCustomShortcode(e.target.value)}
            placeholder="my-custom-link"
          />
        </div>
        <div className="form-group">
          <label htmlFor="validity">Validity (minutes)</label>
          <input
            type="number"
            id="validity"
            value={validity}
            onChange={(e) => setValidity(parseInt(e.target.value, 10))}
            min="1"
            required
          />
        </div>
        <button type="submit" disabled={loading}>
          {loading ? 'Shortening...' : 'Shorten'}
        </button>
      </form>
      {error && <p className="message error-message">{error}</p>}
      {result && (
        <div className="message success-message">
          <p>Success! Your short link is:</p>
          <a href={result.shortLink} target="_blank" rel="noopener noreferrer">
            {result.shortLink}
          </a>
          <p>Expires on: {new Date(result.expiry).toLocaleString()}</p>
        </div>
      )}
    </div>
  );
};

// --- URL Stats Page Component ---
const URLStatsPage: React.FC = () => {
    const [shortUrl, setShortUrl] = useState('');
    const [stats, setStats] = useState<URLStatsResponse | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError(null);
        setStats(null);

        if (!shortUrl) {
            setError('Please enter a short URL to get stats.');
            setLoading(false);
            return;
        }

        try {
            // Extract shortcode from the full URL
            const shortcode = shortUrl.split('/').pop();
            if (!shortcode) {
                throw new Error('Invalid short URL format.');
            }

            const response = await fetch(`${API_BASE_URL}/${shortcode}/stats`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            const data: URLStatsResponse | ApiError = await response.json();

            if (!response.ok) {
                throw new Error((data as ApiError).error || 'Could not fetch stats.');
            }

            setStats(data as URLStatsResponse);
        } catch (err: any) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

  return (
    <div className="card">
      <h2>Get URL Statistics</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="shortUrl">Short URL</label>
          <input
            type="text"
            id="shortUrl"
            value={shortUrl}
            onChange={(e) => setShortUrl(e.target.value)}
            placeholder="http://localhost:8080/abcd1"
            required
          />
        </div>
        <button type="submit" disabled={loading}>
          {loading ? 'Fetching...' : 'Get Stats'}
        </button>
      </form>
      {error && <p className="message error-message">{error}</p>}
      {stats && (
        <div className="message stats-result">
          <h3>Statistics</h3>
          <p><strong>Original URL:</strong> <a href={stats.originalUrl} target="_blank" rel="noopener noreferrer">{stats.originalUrl}</a></p>
          <p><strong>Created At:</strong> {new Date(stats.createdAt).toLocaleString()}</p>
          <p><strong>Expires At:</strong> {new Date(stats.expiresAt).toLocaleString()}</p>
          <p><strong>Total Hits:</strong> {stats.hits}</p>
        </div>
      )}
    </div>
  );
};

export default App;