import { Link } from 'react-router-dom';
import './GetStartedSection.css';

export function GetStartedSection() {
    return (
        <div className="gs-panel">
            <div className="gs-content">
                <h2 className="gs-heading">Stop doing this<br />manually.</h2>
                <p className="gs-sub">
                    Upload one image. Get all 14 Steam assets,<br />pixel-perfect, in 60 seconds.
                </p>
                <Link to="/upload" className="gs-btn">GET STARTED →</Link>
                <p className="gs-note">No account · $9 per pack · Files deleted after 24h</p>
            </div>
        </div>
    );
}
