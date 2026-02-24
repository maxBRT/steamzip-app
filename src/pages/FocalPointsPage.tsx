import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { SiteHeader } from '../components/sections/SiteHeader';
import { ImageCropper } from '../components/ui/ImageCropper';
import { useSession } from '../hooks/useSession';
import { useFocalPoints } from '../hooks/useFocalPoints';
import { STEAM_ASSETS } from '../lib/assets';
import { api } from '../lib/api';
import './FocalPointsPage.css';

export default function FocalPointsPage(): React.ReactElement {
    const session = useSession();
    const fp = useFocalPoints(session.sessionId);
    const [imageUrl, setImageUrl] = useState<string | null>(null);
    const navigate = useNavigate();

    useEffect(() => {
        if (session.status !== 'ready' || !session.resumeState) return;
        if (session.resumeState.paymentStatus === 'paid') {
            navigate('/processing', { replace: true });
        }
    }, [session.status, session.resumeState, session.sessionId, navigate]);

    useEffect(() => {
        if (!session.sessionId) return;
        api.getMasterImageUrl(session.sessionId)
            .then(({ url }) => setImageUrl(url))
            .catch(() => { /* image unavailable, cropper stays hidden */ });
    }, [session.sessionId]);

    async function handleConfirmOrSubmit() {
        if (fp.allConfirmed) {
            if (!session.sessionId) return;
            fp.setSubmitting();
            try {
                const data = await api.createCheckoutSession(session.sessionId, fp.buildPayload());
                window.location.href = data.checkout_url;
            } catch (err) {
                fp.setSubmitError(err instanceof Error ? err.message : 'Failed to create checkout');
            }
        } else {
            fp.confirm();
            if (fp.assetIndex < STEAM_ASSETS.length - 1) {
                fp.goNext();
            }
        }
    }

    const isDone = fp.submitStatus === 'done';

    return (
        <div className="fp-page">
            <SiteHeader />

            <div className="fp-canvas">
                <div className="fp-grid" aria-hidden />
                <div className="fp-watermark" aria-hidden>SZ</div>

                <div className="fp-corner">
                    <span className="fp-corner-num">02</span>
                    <span className="fp-corner-label">Focal Points</span>
                </div>

                {isDone ? (
                    <div className="fp-done">
                        <div className="fp-done-check">✓</div>
                        <div className="fp-done-label">Redirecting to checkout…</div>
                    </div>
                ) : (
                    <div className="fp-content">
                        <div className="fp-nav">
                            <button
                                className="fp-nav-btn"
                                onClick={fp.goPrev}
                                disabled={fp.assetIndex === 0}
                            >
                                ← PREV
                            </button>
                            <div className="fp-nav-center">
                                <span className="fp-nav-name">{fp.currentAsset.name.toUpperCase()}</span>
                                <span className="fp-nav-dim">{fp.currentAsset.width} × {fp.currentAsset.height}</span>
                                <span className="fp-nav-count">[{fp.assetIndex + 1} / {STEAM_ASSETS.length}]</span>
                            </div>
                            <button
                                className="fp-nav-btn"
                                onClick={fp.goNext}
                                disabled={fp.assetIndex === STEAM_ASSETS.length - 1}
                            >
                                NEXT →
                            </button>
                        </div>

                        {imageUrl && (
                            <ImageCropper
                                imageUrl={imageUrl}
                                asset={fp.currentAsset}
                                point={fp.focalPoints[fp.currentAsset.id] ?? { x: 0.5, y: 0.5 }}
                                zoom={fp.zooms[fp.currentAsset.id] ?? 1}
                                onChange={fp.setPoint}
                                onZoomChange={fp.setZoom}
                            />
                        )}

                        <div className="fp-progress">
                            {STEAM_ASSETS.map((asset, i) => (
                                <button
                                    key={asset.id}
                                    className={[
                                        'fp-progress-sq',
                                        fp.confirmedIds.has(asset.id) ? 'fp-progress-sq--confirmed' : '',
                                        i === fp.assetIndex ? 'fp-progress-sq--active' : '',
                                    ].join(' ').trim()}
                                    onClick={() => fp.goTo(i)}
                                    title={asset.name}
                                    aria-label={asset.name}
                                />
                            ))}
                        </div>

                        <button
                            className="fp-submit"
                            onClick={handleConfirmOrSubmit}
                            disabled={fp.submitStatus === 'submitting'}
                        >
                            {fp.submitStatus === 'submitting'
                                ? 'SUBMITTING…'
                                : fp.allConfirmed
                                    ? 'SUBMIT ALL →'
                                    : 'CONFIRM CROP →'}
                        </button>

                        {fp.submitError && (
                            <div className="fp-error">{fp.submitError}</div>
                        )}

                        <button
                            className="fp-change-image"
                            onClick={() => navigate('/upload?reupload=true')}
                        >
                            ← CHANGE IMAGE
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
