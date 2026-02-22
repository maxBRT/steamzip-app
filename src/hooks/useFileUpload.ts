import { useState } from 'react';
import { api } from '../lib/api';

export type UploadStatus = 'idle' | 'uploading' | 'done' | 'error';

export interface UseFileUploadReturn {
    upload: (file: File) => Promise<void>;
    status: UploadStatus;
    error: string;
}

export function useFileUpload(sessionId: string | null): UseFileUploadReturn {
    const [status, setStatus] = useState<UploadStatus>('idle');
    const [error, setError] = useState<string>('');

    async function upload(file: File): Promise<void> {
        if (!sessionId) return;
        setStatus('uploading');
        setError('');
        try {
            const { uploadUrl } = await api.getUploadUrl(sessionId, file.type);
            await fetch(uploadUrl, { method: 'PUT', body: file, headers: { 'Content-Type': file.type } });
            await api.confirmUpload(sessionId);
            setStatus('done');
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Upload failed');
            setStatus('error');
        }
    }

    return { upload, status, error };
}
