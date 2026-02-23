import { useState } from 'react';
import { STEAM_ASSETS, type SteamAsset } from '../lib/assets';
import { api, type FocalPointsPayload } from '../lib/api';

export interface FocalPoint { x: number; y: number; }
export type FocalPointsMap = Record<string, FocalPoint>;
export type ZoomsMap = Record<string, number>;
export type SubmitStatus = 'idle' | 'submitting' | 'done' | 'error';

export interface UseFocalPointsReturn {
    assetIndex: number;
    currentAsset: SteamAsset;
    focalPoints: FocalPointsMap;
    zooms: ZoomsMap;
    confirmedIds: Set<string>;
    allConfirmed: boolean;
    setPoint: (point: FocalPoint) => void;
    setZoom: (zoom: number) => void;
    confirm: () => void;
    goTo: (index: number) => void;
    goNext: () => void;
    goPrev: () => void;
    submitStatus: SubmitStatus;
    submitError: string;
    submitAll: () => Promise<void>;
    buildPayload: () => FocalPointsPayload;
    setSubmitting: () => void;
    setSubmitError: (err: string) => void;
}

const INITIAL_POINTS: FocalPointsMap = Object.fromEntries(
    STEAM_ASSETS.map(a => [a.id, { x: 0.5, y: 0.5 }])
);

const INITIAL_ZOOMS: ZoomsMap = Object.fromEntries(
    STEAM_ASSETS.map(a => [a.id, 1])
);

export function useFocalPoints(sessionId: string | null): UseFocalPointsReturn {
    const [assetIndex, setAssetIndex] = useState(0);
    const [focalPoints, setFocalPoints] = useState<FocalPointsMap>(INITIAL_POINTS);
    const [zooms, setZoomsMap] = useState<ZoomsMap>(INITIAL_ZOOMS);
    const [confirmedIds, setConfirmedIds] = useState<Set<string>>(new Set());
    const [submitStatus, setSubmitStatus] = useState<SubmitStatus>('idle');
    const [submitError, setSubmitError] = useState('');

    const currentAsset = STEAM_ASSETS[assetIndex]!;
    const allConfirmed = confirmedIds.size === STEAM_ASSETS.length;

    function setPoint(point: FocalPoint): void {
        setFocalPoints(prev => ({ ...prev, [currentAsset.id]: point }));
    }

    function setZoom(zoom: number): void {
        setZoomsMap(prev => ({ ...prev, [currentAsset.id]: Math.min(4, Math.max(1, zoom)) }));
    }

    function confirm(): void {
        setConfirmedIds(prev => new Set([...prev, currentAsset.id]));
    }

    function goTo(index: number): void {
        if (index >= 0 && index < STEAM_ASSETS.length) setAssetIndex(index);
    }

    function goNext(): void { goTo(assetIndex + 1); }
    function goPrev(): void { goTo(assetIndex - 1); }

    function buildPayload(): FocalPointsPayload {
        // Backend expects snake_case keys; convert from internal hyphen-case
        return Object.fromEntries(
            STEAM_ASSETS.map(a => [a.id.replace(/-/g, '_'), { ...focalPoints[a.id]!, zoom: zooms[a.id]! }])
        );
    }

    function setSubmitting(): void {
        setSubmitStatus('submitting');
        setSubmitError('');
    }

    function setSubmitErrorFn(err: string): void {
        setSubmitError(err);
        setSubmitStatus('error');
    }

    async function submitAll(): Promise<void> {
        if (!sessionId) return;
        setSubmitting();
        try {
            await api.submitFocalPoints(sessionId, buildPayload());
            setSubmitStatus('done');
        } catch (err: unknown) {
            setSubmitErrorFn(err instanceof Error ? err.message : 'Submit failed');
        }
    }

    return {
        assetIndex,
        currentAsset,
        focalPoints,
        zooms,
        confirmedIds,
        allConfirmed,
        setPoint,
        setZoom,
        confirm,
        goTo,
        goNext,
        goPrev,
        submitStatus,
        submitError,
        submitAll,
        buildPayload,
        setSubmitting,
        setSubmitError: setSubmitErrorFn,
    };
}
