import { useState } from 'react';
import { STEAM_ASSETS, type SteamAsset } from '../lib/assets';
import { api, type FocalPointsPayload } from '../lib/api';

export interface FocalPoint { x: number; y: number; }
export type FocalPointsMap = Record<string, FocalPoint>;
export type SubmitStatus = 'idle' | 'submitting' | 'done' | 'error';

export interface UseFocalPointsReturn {
    assetIndex: number;
    currentAsset: SteamAsset;
    focalPoints: FocalPointsMap;
    confirmedIds: Set<string>;
    allConfirmed: boolean;
    setPoint: (point: FocalPoint) => void;
    confirm: () => void;
    goTo: (index: number) => void;
    goNext: () => void;
    goPrev: () => void;
    submitStatus: SubmitStatus;
    submitError: string;
    submitAll: () => Promise<void>;
}

const INITIAL_POINTS: FocalPointsMap = Object.fromEntries(
    STEAM_ASSETS.map(a => [a.id, { x: 0.5, y: 0.5 }])
);

export function useFocalPoints(sessionId: string | null): UseFocalPointsReturn {
    const [assetIndex, setAssetIndex] = useState(0);
    const [focalPoints, setFocalPoints] = useState<FocalPointsMap>(INITIAL_POINTS);
    const [confirmedIds, setConfirmedIds] = useState<Set<string>>(new Set());
    const [submitStatus, setSubmitStatus] = useState<SubmitStatus>('idle');
    const [submitError, setSubmitError] = useState('');

    const currentAsset = STEAM_ASSETS[assetIndex];
    const allConfirmed = confirmedIds.size === STEAM_ASSETS.length;

    function setPoint(point: FocalPoint): void {
        setFocalPoints(prev => ({ ...prev, [currentAsset.id]: point }));
    }

    function confirm(): void {
        setConfirmedIds(prev => new Set([...prev, currentAsset.id]));
    }

    function goTo(index: number): void {
        if (index >= 0 && index < STEAM_ASSETS.length) setAssetIndex(index);
    }

    function goNext(): void { goTo(assetIndex + 1); }
    function goPrev(): void { goTo(assetIndex - 1); }

    async function submitAll(): Promise<void> {
        if (!sessionId) return;
        setSubmitStatus('submitting');
        setSubmitError('');
        try {
            // Backend expects { focalPoints: { "header_capsule": {x, y}, ... } }
            // Asset IDs use hyphens internally; convert to underscores for the API.
            const points: FocalPointsPayload = Object.fromEntries(
                STEAM_ASSETS.map(a => [a.id.replace(/-/g, '_'), focalPoints[a.id]])
            );
            await api.submitFocalPoints(sessionId, points);
            setSubmitStatus('done');
        } catch (err: unknown) {
            setSubmitError(err instanceof Error ? err.message : 'Submit failed');
            setSubmitStatus('error');
        }
    }

    return {
        assetIndex,
        currentAsset,
        focalPoints,
        confirmedIds,
        allConfirmed,
        setPoint,
        confirm,
        goTo,
        goNext,
        goPrev,
        submitStatus,
        submitError,
        submitAll,
    };
}
