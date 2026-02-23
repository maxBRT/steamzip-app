import { BrowserRouter, Routes, Route } from 'react-router-dom';
import LandingPage from "./pages/LandingPage";
import UploadPage from "./pages/UploadPage";
import FocalPointsPage from "./pages/FocalPointsPage";
import ProcessingPage from "./pages/ProcessingPage";

export function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<LandingPage />} />
                <Route path="/upload" element={<UploadPage />} />
                <Route path="/focal-points" element={<FocalPointsPage />} />
                <Route path="/processing" element={<ProcessingPage />} />
            </Routes>
        </BrowserRouter>
    );
}

export default App;
