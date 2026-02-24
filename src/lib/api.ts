const BASE = import.meta.env?.BUN_PUBLIC_API_URL ?? '';

export class ApiError extends Error {
    constructor(public status: number, message: string) {
        super(message);
    }
}

export interface CreateSessionResponse {
    session_id: string;
}

export interface GetUploadUrlResponse {
    assetId: string;
    uploadUrl: string;
}

export interface GetMasterImageUrlResponse {
    url: string;
}

export type FocalPointsPayload = Record<string, { x: number; y: number; zoom: number }>;

export interface CreateCheckoutResponse {
    checkout_url: string;
}

export interface GetSessionStatusResponse {
    sessionId: string;
    status: string;
    assetStatus?: string;
    downloadUrl?: string;
}

interface Api {
    createSession: () => Promise<CreateSessionResponse>;
    getUploadUrl: (sessionId: string, contentType: string) => Promise<GetUploadUrlResponse>;
    confirmUpload: (sessionId: string) => Promise<void>;
    getMasterImageUrl: (sessionId: string) => Promise<GetMasterImageUrlResponse>;
    submitFocalPoints: (sessionId: string, points: FocalPointsPayload) => Promise<void>;
    createCheckoutSession: (sessionId: string, focalPoints: FocalPointsPayload) => Promise<CreateCheckoutResponse>;
    getSessionStatus: (sessionId: string) => Promise<GetSessionStatusResponse>;
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
    const res = await fetch(`${BASE}${path}`, {
        headers: { 'Content-Type': 'application/json', ...init?.headers },
        ...init,
    });
    if (!res.ok) throw new ApiError(res.status, `API ${res.status}: ${await res.text()}`);
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
    getMasterImageUrl: (sessionId: string) =>
        apiFetch<GetMasterImageUrlResponse>(`/api/sessions/${sessionId}/assets/master`),
    submitFocalPoints: (sessionId: string, points: FocalPointsPayload) =>
        apiFetch<void>(`/api/sessions/${sessionId}/assets/process`, {
            method: 'POST',
            body: JSON.stringify({ focalPoints: points }),
        }),
    createCheckoutSession: (sessionId: string, focalPoints: FocalPointsPayload) =>
        apiFetch<CreateCheckoutResponse>(`/api/sessions/${sessionId}/checkout`, {
            method: 'POST',
            body: JSON.stringify({ focal_points: focalPoints }),
        }),
    getSessionStatus: (sessionId: string) =>
        apiFetch<GetSessionStatusResponse>(`/api/sessions/${sessionId}/status`),
};
