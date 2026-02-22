import { SiteHeader } from '../components/sections/SiteHeader';
import { LeftPanel } from '../components/sections/LeftPanel';
import { FeaturesGrid } from '../components/sections/FeaturesGrid';
import { AssetsMarquee } from '../components/sections/AssetsMarquee';
import { GetStartedSection } from '../components/sections/GetStartedSection';
import { SiteFooter } from '../components/sections/SiteFooter';
import './LandingPage.css';

export default function LandingPage() {
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
                <GetStartedSection />
            </div>
            <SiteFooter />
        </div>
    );
}
