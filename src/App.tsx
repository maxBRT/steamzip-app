import { BrowserRouter, Routes, Route } from 'react-router-dom';
import LandingPage from "./pages/LandingPage";
import UploadPage from "./pages/UploadPage";
import FocalPointsPage from "./pages/FocalPointsPage";

export function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<LandingPage />} />
                <Route path="/upload" element={<UploadPage />} />
                <Route path="/focal-points" element={<FocalPointsPage />} />
            </Routes>
        </BrowserRouter>
    );
}

export default App;
