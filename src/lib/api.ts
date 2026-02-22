const BASE = import.meta.env?.BUN_PUBLIC_API_URL ?? '';

export interface CreateSessionResponse {
    session_id: string;
}

export interface GetUploadUrlResponse {
    assetId: string;
    uploadUrl: string;
}

interface Api {
    createSession: () => Promise<CreateSessionResponse>;
    getUploadUrl: (sessionId: string, contentType: string) => Promise<GetUploadUrlResponse>;
    confirmUpload: (sessionId: string) => Promise<void>;
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
    const res = await fetch(`${BASE}${path}`, {
        headers: { 'Content-Type': 'application/json', ...init?.headers },
        ...init,
    });
    if (!res.ok) throw new Error(`API ${res.status}: ${await res.text()}`);
    if (res.status === 204) return undefined as T;
    return res.json() as Promise<T>;
}

export const api: Api = {
    createSession: () => apiFetch<CreateSessionResponse>('/api/sessions', { method: 'POST' }),
    getUploadUrl: (sessionId: string, contentType: string) =>
        apiFetch<GetUploadUrlResponse>(
            `/api/sessions/${sessionId}/assets/upload-url`,
            { method: 'POST', body: JSON.stringify({ contentType }) }
        ),
    confirmUpload: (sessionId: string) =>
        apiFetch<void>(`/api/sessions/${sessionId}/assets/confirm-upload`, { method: 'POST' }),
};
