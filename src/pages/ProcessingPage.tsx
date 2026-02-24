import { useEffect, useRef, useState } from 'react';
import { SiteHeader } from '../components/sections/SiteHeader';
import { api } from '../lib/api';
import './ProcessingPage.css';

type PageState = 'processing' | 'ready' | 'error';

export default function ProcessingPage(): React.ReactElement {
    const sessionId = localStorage.getItem('sz_session_id');

    const [pageState, setPageState] = useState<PageState>('processing');
    const [downloadUrl, setDownloadUrl] = useState<string | null>(null);
    const [errorMsg, setErrorMsg] = useState('');
    const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

    useEffect(() => {
        if (!sessionId) {
            setErrorMsg('No active session found.');
            setPageState('error');
            return;
        }

        function poll() {
            api.getSessionStatus(sessionId!)
                .then(data => {
                    if (data.assetStatus === 'ready' && data.downloadUrl) {
                        setDownloadUrl(data.downloadUrl);
                        setPageState('ready');
                        if (intervalRef.current) clearInterval(intervalRef.current);
                    } else if (data.assetStatus === 'failed') {
                        setErrorMsg('Processing failed. Please try again.');
                        setPageState('error');
                        if (intervalRef.current) clearInterval(intervalRef.current);
                    }
                })
                .catch(() => {
                    // Keep polling; transient network errors shouldn't stop us
                });
        }

        poll();
        intervalRef.current = setInterval(poll, 3000);

        return () => {
            if (intervalRef.current) clearInterval(intervalRef.current);
        };
    }, [sessionId]);

    return (
        <div className="pp-page">
            <SiteHeader />

            <div className="pp-canvas">
                <div className="pp-grid" aria-hidden />
                <div className="pp-watermark" aria-hidden>SZ</div>

                <div className="pp-corner">
                    <span className="pp-corner-num">03</span>
                    <span className="pp-corner-label">Processing</span>
                </div>

                {pageState === 'processing' && (
                    <div className="pp-content">
                        <div className="pp-spinner" aria-label="Processing…" />
                        <div className="pp-label">Processing.</div>
                        <div className="pp-sub">Cropping &amp; resizing your Steam assets…</div>
                    </div>
                )}

                {pageState === 'ready' && downloadUrl && (
                    <div className="pp-content">
                        <div className="pp-check">✓</div>
                        <div className="pp-label">Done.</div>
                        <a
                            className="pp-download"
                            href={downloadUrl}
                            download
                        >
                            Download ZIP →
                        </a>
                    </div>
                )}

                {pageState === 'error' && (
                    <div className="pp-content">
                        <div className="pp-error-label">Error.</div>
                        <div className="pp-sub">{errorMsg}</div>
                    </div>
                )}
            </div>
        </div>
    );
}
