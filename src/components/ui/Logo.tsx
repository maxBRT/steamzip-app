import { useNavigate } from 'react-router-dom';
import './Logo.css';

export function Logo() {
    const navigate = useNavigate();
    return (
        <div className="d5-logo" onClick={() => navigate('/')} role="link" tabIndex={0} onKeyDown={(e) => e.key === 'Enter' && navigate('/')}>
            Steam<span>Zip</span>
        </div>
    );
}
