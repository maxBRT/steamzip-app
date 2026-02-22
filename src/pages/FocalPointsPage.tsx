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
    const navigate = useNavigate();
    const [imageUrl, setImageUrl] = useState<string | null>(null);

    useEffect(() => {
        if (!session.sessionId) return;
        api.getMasterImageUrl(session.sessionId)
            .then(({ url }) => setImageUrl(url))
            .catch(() => { /* image unavailable, cropper stays hidden */ });
    }, [session.sessionId]);

    useEffect(() => {
        if (fp.submitStatus === 'done') {
            // Future: navigate to a results page
        }
    }, [fp.submitStatus, navigate]);

    function handleConfirmOrSubmit() {
        if (fp.allConfirmed) {
            fp.submitAll();
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
                        <div className="fp-done-label">Processing.</div>
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
                                onChange={fp.setPoint}
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
                    </div>
                )}
            </div>
        </div>
    );
}
