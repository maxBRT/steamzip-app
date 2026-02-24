import './GetStartedSection.css';

interface Props {
    onGetStarted: () => void;
    checking: boolean;
}

export function GetStartedSection({ onGetStarted, checking }: Props) {
    return (
        <div className="gs-panel">
            <div className="gs-content">
                <h2 className="gs-heading">Stop doing this<br />manually.</h2>
                <p className="gs-sub">
                    Upload one image. Get all 14 Steam assets,<br />pixel-perfect, in 60 seconds.
                </p>
                <button
                    className="gs-btn"
                    onClick={onGetStarted}
                    disabled={checking}
                >
                    {checking ? 'LOADING…' : 'GET STARTED →'}
                </button>
                <p className="gs-note">No account · $9 per pack · Files deleted after 24h</p>
            </div>
        </div>
    );
}
