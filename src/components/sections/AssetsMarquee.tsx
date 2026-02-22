import { STEAM_ASSETS } from '../../lib/assets';
import './AssetsMarquee.css';

export function AssetsMarquee() {
    const items = [...STEAM_ASSETS, ...STEAM_ASSETS];
    return (
        <div className="d5-marquee">
            <div className="d5-marquee-track">
                {items.map((a, i) => (
                    <span key={i} className="d5-marquee-item">
                        <span className="d5-marquee-name">{a.name}</span>
                        <span className="d5-marquee-dim">{`${a.width} × ${a.height}`}</span>
                        <span className="d5-marquee-sep">◆</span>
                    </span>
                ))}
            </div>
        </div>
    );
}
