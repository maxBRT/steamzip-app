import { useRef, useEffect, useState } from 'react';
import type { SteamAsset } from '../../lib/assets';
import type { FocalPoint } from '../../hooks/useFocalPoints';
import './ImageCropper.css';

interface ImageCropperProps {
    imageUrl: string;
    asset: SteamAsset;
    point: FocalPoint;
    onChange: (p: FocalPoint) => void;
}

export function ImageCropper({ imageUrl, asset, point, onChange }: ImageCropperProps) {
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
    // object-fit: contain scales the image uniformly to fit, adding bars on the short axis.
    const { width: dw, height: dh } = displaySize;
    const { width: nw, height: nh } = naturalSize;

    let imgX = 0, imgY = 0, imgW = 0, imgH = 0;
    if (nw > 0 && nh > 0 && dw > 0 && dh > 0) {
        const containerAR = dw / dh;
        const imageAR = nw / nh;
        if (imageAR > containerAR) {
            // Image wider than container slot → fit to width, bars on top/bottom
            imgW = dw;
            imgH = dw / imageAR;
            imgX = 0;
            imgY = (dh - imgH) / 2;
        } else {
            // Image taller than container slot → fit to height, bars on left/right
            imgH = dh;
            imgW = dh * imageAR;
            imgX = (dw - imgW) / 2;
            imgY = 0;
        }
    }

    // Scale factor: display pixels per natural pixel.
    const scale = imgW > 0 ? imgW / nw : 0;

    // Mirror backend CropRect: the crop region on the master is fitted to the target
    // aspect ratio, constrained by whichever master dimension is the bottleneck.
    //   masterAR > targetAR → master is wider → crop fits to master HEIGHT
    //   masterAR ≤ targetAR → master is taller → crop fits to master WIDTH
    const targetAR = asset.width / asset.height;
    const masterAR = nw > 0 && nh > 0 ? nw / nh : 1;
    const cropNatW = masterAR > targetAR ? nh * targetAR : nw;
    const cropNatH = masterAR > targetAR ? nh : nw / targetAR;
    const rectW = cropNatW * scale;
    const rectH = cropNatH * scale;

    // Rect top-left in container-space, clamped within the rendered image area.
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

        // Clamp rect top-left within the rendered image area.
        const cl = Math.max(imgX, Math.min(pointerX - dragOffset.current.dx, imgX + imgW - rectW));
        const ct = Math.max(imgY, Math.min(pointerY - dragOffset.current.dy, imgY + imgH - rectH));

        // Normalize center coords relative to the rendered image, matching backend expectations.
        onChange({
            x: (cl - imgX + rectW / 2) / imgW,
            y: (ct - imgY + rectH / 2) / imgH,
        });
    }

    function handlePointerUp(e: React.PointerEvent<HTMLDivElement>) {
        e.currentTarget.releasePointerCapture(e.pointerId);
        dragOffset.current = null;
    }

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
        </div>
    );
}
