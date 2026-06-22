import { Component, type ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

export default class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { error, errorInfo: null };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({ error, errorInfo });
    // eslint-disable-next-line no-console
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  render() {
    if (this.state.error) {
      return (
        <div style={{
          padding: '2rem',
          fontFamily: 'system-ui, sans-serif',
          backgroundColor: 'var(--color-base-background)',
          color: 'var(--color-base-font)',
          minHeight: '100vh',
        }}>
          <h1 style={{ fontSize: '1.5rem', marginBottom: '1rem', color: 'var(--color-engine-error)' }}>
            Something went wrong
          </h1>
          <p style={{ marginBottom: '1rem' }}>
            The application crashed. Please refresh the page or try again later.
          </p>
          <details style={{ marginBottom: '1rem' }}>
            <summary style={{ cursor: 'pointer', fontWeight: 600 }}>Error details</summary>
            <pre style={{
              marginTop: '0.5rem',
              padding: '1rem',
              backgroundColor: 'var(--color-result-background)',
              border: '1px solid var(--color-result-border)',
              borderRadius: '0.25rem',
              overflow: 'auto',
              fontSize: '0.85rem',
            }}>
              {this.state.error.toString()}
              {'\n'}
              {this.state.errorInfo?.componentStack}
            </pre>
          </details>
          <button
            onClick={() => window.location.reload()}
            style={{
              padding: '0.5rem 1rem',
              backgroundColor: 'var(--color-button-background)',
              color: 'var(--color-button-font)',
              border: 'none',
              borderRadius: '0.25rem',
              cursor: 'pointer',
            }}
          >
            Reload page
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}
