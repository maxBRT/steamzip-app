import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { SiteHeader } from '../components/sections/SiteHeader';
import { DropZone } from '../components/ui/DropZone';
import { useSession } from '../hooks/useSession';
import { useFileUpload } from '../hooks/useFileUpload';
import './UploadPage.css';

export default function UploadPage(): React.ReactElement {
    const session = useSession();
    const { upload, status: uploadStatus, error: uploadError } = useFileUpload(session.sessionId);
    const [file, setFile] = useState<File | null>(null);
    const navigate = useNavigate();

    const isLoading   = session.status === 'loading';
    const isUploading = uploadStatus === 'uploading';
    const isError     = session.status === 'error' || uploadStatus === 'error';
    const error       = session.error || uploadError;

    useEffect(() => {
        if (uploadStatus === 'done') {
            navigate('/focal-points');
        }
    }, [uploadStatus, navigate]);

    function handleSubmit(e: React.FormEvent<HTMLFormElement>): void {
        e.preventDefault();
        if (file) upload(file);
    }

    return (
        <div className="up-page">
            <SiteHeader />

            <div className="up-canvas">
                <div className="up-grid" aria-hidden />
                <div className="up-watermark" aria-hidden>SZ</div>

                <div className="up-corner">
                    <span className="up-corner-num">01</span>
                    <span className="up-corner-label">Master Image</span>
                </div>

                {isError && (
                    <div className="up-error">
                        <div className="up-error-label">Error</div>
                        <div className="up-error-msg">{error}</div>
                        <button className="up-retry" onClick={session.reset}>Retry</button>
                    </div>
                )}

                {!isError && (
                    <form
                        className={`up-form${isLoading ? ' up-form--init' : ''}`}
                        onSubmit={handleSubmit}
                    >
                        <DropZone
                            file={file}
                            disabled={isLoading || isUploading}
                            onChange={setFile}
                        />
                        <button
                            type="submit"
                            className="up-submit"
                            disabled={!file || isLoading || isUploading}
                        >
                            {isUploading ? 'UPLOADING…' : 'UPLOAD →'}
                        </button>
                    </form>
                )}
            </div>
        </div>
    );
}
