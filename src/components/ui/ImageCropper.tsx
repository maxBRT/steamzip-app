import { useRef, useEffect, useState } from 'react';
import type { SteamAsset } from '../../lib/assets';
import type { FocalPoint } from '../../hooks/useFocalPoints';
import './ImageCropper.css';

const MIN_ZOOM = 1;
const MAX_ZOOM = 4;
const ZOOM_STEP = 0.25;

interface ImageCropperProps {
    imageUrl: string;
    asset: SteamAsset;
    point: FocalPoint;
    zoom: number;
    onChange: (p: FocalPoint) => void;
    onZoomChange: (z: number) => void;
}

export function ImageCropper({ imageUrl, asset, point, zoom, onChange, onZoomChange }: ImageCropperProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const [displaySize, setDisplaySize] = useState({ width: 0, height: 0 });
    const [naturalSize, setNaturalSize] = useState({ width: 0, height: 0 });
    const dragOffset = useRef<{ dx: number; dy: number } | null>(null);

    useEffect(() => {
        const el = containerRef.current;
        if (!el) return;
        const ro = new ResizeObserver(entries => {
            for (const entry of entries) {
                const { width, height } = entry.contentRect;
                setDisplaySize({ width, height });
            }
        });
        ro.observe(el);
        return () => ro.disconnect();
    }, []);

    // Compute the letterboxed image rect inside the container.
    const { width: dw, height: dh } = displaySize;
    const { width: nw, height: nh } = naturalSize;

    let imgX = 0, imgY = 0, imgW = 0, imgH = 0;
    if (nw > 0 && nh > 0 && dw > 0 && dh > 0) {
        const containerAR = dw / dh;
        const imageAR = nw / nh;
        if (imageAR > containerAR) {
            imgW = dw;
            imgH = dw / imageAR;
            imgX = 0;
            imgY = (dh - imgH) / 2;
        } else {
            imgH = dh;
            imgW = dh * imageAR;
            imgX = (dw - imgW) / 2;
            imgY = 0;
        }
    }

    const scale = imgW > 0 ? imgW / nw : 0;

    const targetAR = asset.width / asset.height;
    const masterAR = nw > 0 && nh > 0 ? nw / nh : 1;
    const cropNatW = masterAR > targetAR ? nh * targetAR : nw;
    const cropNatH = masterAR > targetAR ? nh : nw / targetAR;

    // Zoom shrinks the crop rect — zoom=2 means half the area on each axis.
    const rectW = (cropNatW * scale) / zoom;
    const rectH = (cropNatH * scale) / zoom;

    const clampedLeft = Math.max(imgX, Math.min(imgX + point.x * imgW - rectW / 2, imgX + imgW - rectW));
    const clampedTop = Math.max(imgY, Math.min(imgY + point.y * imgH - rectH / 2, imgY + imgH - rectH));

    function handlePointerDown(e: React.PointerEvent<HTMLDivElement>) {
        e.preventDefault();
        e.currentTarget.setPointerCapture(e.pointerId);
        const containerRect = containerRef.current!.getBoundingClientRect();
        dragOffset.current = {
            dx: (e.clientX - containerRect.left) - clampedLeft,
            dy: (e.clientY - containerRect.top) - clampedTop,
        };
    }

    function handlePointerMove(e: React.PointerEvent<HTMLDivElement>) {
        if (!dragOffset.current || !containerRef.current) return;
        const containerRect = containerRef.current.getBoundingClientRect();
        const pointerX = e.clientX - containerRect.left;
        const pointerY = e.clientY - containerRect.top;

        const cl = Math.max(imgX, Math.min(pointerX - dragOffset.current.dx, imgX + imgW - rectW));
        const ct = Math.max(imgY, Math.min(pointerY - dragOffset.current.dy, imgY + imgH - rectH));

        onChange({
            x: (cl - imgX + rectW / 2) / imgW,
            y: (ct - imgY + rectH / 2) / imgH,
        });
    }

    function handlePointerUp(e: React.PointerEvent<HTMLDivElement>) {
        e.currentTarget.releasePointerCapture(e.pointerId);
        dragOffset.current = null;
    }

    useEffect(() => {
        const el = containerRef.current;
        if (!el) return;
        function onWheel(e: WheelEvent) {
            e.preventDefault();
            const direction = e.deltaY > 0 ? 1 : -1;
            const newZoom = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, zoom + direction * ZOOM_STEP));
            onZoomChange(Math.round(newZoom / ZOOM_STEP) * ZOOM_STEP);
        }
        el.addEventListener('wheel', onWheel, { passive: false });
        return () => el.removeEventListener('wheel', onWheel);
    }, [zoom, onZoomChange]);

    const ready = imgW > 0 && rectW > 0;

    return (
        <div className="ic-container" ref={containerRef}>
            <img
                className="ic-image"
                src={imageUrl}
                alt="Master"
                draggable={false}
                onLoad={e => setNaturalSize({
                    width: e.currentTarget.naturalWidth,
                    height: e.currentTarget.naturalHeight,
                })}
            />
            {ready && (
                <div
                    className="ic-rect"
                    style={{ left: clampedLeft, top: clampedTop, width: rectW, height: rectH }}
                    onPointerDown={handlePointerDown}
                    onPointerMove={handlePointerMove}
                    onPointerUp={handlePointerUp}
                />
            )}
            <div className="ic-zoom">
                <button
                    className="ic-zoom-btn"
                    onClick={() => onZoomChange(Math.max(MIN_ZOOM, zoom - ZOOM_STEP))}
                    disabled={zoom <= MIN_ZOOM}
                    aria-label="Zoom out"
                >−</button>
                <span className="ic-zoom-label">{zoom.toFixed(2)}×</span>
                <button
                    className="ic-zoom-btn"
                    onClick={() => onZoomChange(Math.min(MAX_ZOOM, zoom + ZOOM_STEP))}
                    disabled={zoom >= MAX_ZOOM}
                    aria-label="Zoom in"
                >+</button>
            </div>
        </div>
    );
}
