import { useState, useEffect } from 'react';
import { api } from '../lib/api';

const SESSION_KEY = 'sz_session_id';

export type SessionStatus = 'loading' | 'ready' | 'error';

export interface UseSessionReturn {
    sessionId: string | null;
    status: SessionStatus;
    error: string;
    reset: () => void;
}

export function useSession(): UseSessionReturn {
    const [sessionId, setSessionId] = useState<string | null>(null);
    const [status, setStatus] = useState<SessionStatus>('loading');
    const [error, setError] = useState<string>('');

    useEffect(() => {
        const existing = localStorage.getItem(SESSION_KEY);
        if (existing) {
            setSessionId(existing);
            setStatus('ready');
            return;
        }
        api.createSession()
            .then(({ session_id }) => {
                localStorage.setItem(SESSION_KEY, session_id);
                setSessionId(session_id);
                setStatus('ready');
            })
            .catch((err: unknown) => {
                setError(err instanceof Error ? err.message : 'Failed to start session');
                setStatus('error');
            });
    }, []);

    function reset(): void {
        localStorage.removeItem(SESSION_KEY);
        window.location.reload();
    }

    return { sessionId, status, error, reset };
}
