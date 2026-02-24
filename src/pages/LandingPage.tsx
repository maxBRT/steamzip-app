import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { SiteHeader } from '../components/sections/SiteHeader';
import { LeftPanel } from '../components/sections/LeftPanel';
import { FeaturesGrid } from '../components/sections/FeaturesGrid';
import { AssetsMarquee } from '../components/sections/AssetsMarquee';
import { GetStartedSection } from '../components/sections/GetStartedSection';
import { SiteFooter } from '../components/sections/SiteFooter';
import { api, ApiError } from '../lib/api';
import { SESSION_KEY } from '../hooks/useSession';
import './LandingPage.css';

export default function LandingPage() {
    const navigate = useNavigate();
    const [checking, setChecking] = useState(false);

    async function handleGetStarted() {
        const existing = localStorage.getItem(SESSION_KEY);
        if (!existing) {
            navigate('/upload');
            return;
        }

        setChecking(true);
        try {
            const data = await api.getSessionStatus(existing);
            const { status: paymentStatus, assetStatus, downloadUrl } = data;
            if (downloadUrl || assetStatus === 'ready' || paymentStatus === 'paid') {
                navigate(`/processing?session=${existing}`);
            } else if (assetStatus === 'uploaded') {
                navigate('/focal-points');
            } else {
                // uploading, undefined, or any other state → go to upload
                navigate('/upload');
            }
        } catch (err) {
            if (err instanceof ApiError && err.status === 410) {
                localStorage.removeItem(SESSION_KEY);
            }
            navigate('/upload');
        }
    }

    return (
        <div className="d5">
            <SiteHeader />
            <div className="d5-top">
                <LeftPanel />
                <FeaturesGrid />
            </div>
            <AssetsMarquee />
            <div className="d5-bottom">
                <div className="d5-bottom-left">
                    <p className="d5-tagline">
                        Steam requires 14 assets with exact pixel dimensions. Upload One Image and get every file, perfectly cropped, in 60 seconds.
                    </p>
                </div>
                <GetStartedSection onGetStarted={handleGetStarted} checking={checking} />
            </div>
            <SiteFooter />
        </div>
    );
}
