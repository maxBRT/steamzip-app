import { useState, useEffect } from 'react';
import { api, ApiError } from '../lib/api';

export const SESSION_KEY = 'sz_session_id';

export type SessionStatus = 'loading' | 'resuming' | 'ready' | 'error';

export interface ResumeState {
    paymentStatus: string;   // 'payment_pending' | 'paid'
    assetStatus?: string;    // 'uploading' | 'uploaded' | 'processing' | 'ready' | 'failed'
    downloadUrl?: string;
}

export interface UseSessionReturn {
    sessionId: string | null;
    status: SessionStatus;
    error: string;
    resumeState: ResumeState | null;
    reset: () => void;
}

export function useSession(): UseSessionReturn {
    const [sessionId, setSessionId] = useState<string | null>(null);
    const [status, setStatus] = useState<SessionStatus>('loading');
    const [error, setError] = useState<string>('');
    const [resumeState, setResumeState] = useState<ResumeState | null>(null);

    useEffect(() => {
        const existing = localStorage.getItem(SESSION_KEY);
        if (existing) {
            setStatus('resuming');
            api.getSessionStatus(existing)
                .then((data) => {
                    setSessionId(existing);
                    setResumeState({
                        paymentStatus: data.status,
                        assetStatus: data.assetStatus,
                        downloadUrl: data.downloadUrl,
                    });
                    setStatus('ready');
                })
                .catch((err: unknown) => {
                    if (err instanceof ApiError && err.status === 410) {
                        // Session expired — clear and create a new one
                        localStorage.removeItem(SESSION_KEY);
                        api.createSession()
                            .then(({ session_id }) => {
                                localStorage.setItem(SESSION_KEY, session_id);
                                setSessionId(session_id);
                                setStatus('ready');
                            })
                            .catch((createErr: unknown) => {
                                setError(createErr instanceof Error ? createErr.message : 'Failed to start session');
                                setStatus('error');
                            });
                    } else {
                        setError(err instanceof Error ? err.message : 'Failed to resume session');
                        setStatus('error');
                    }
                });
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

    return { sessionId, status, error, resumeState, reset };
}
