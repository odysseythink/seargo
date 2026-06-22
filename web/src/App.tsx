import { BrowserRouter, Routes, Route, useLocation } from "react-router-dom";
import LayoutShell from "./components/layout/LayoutShell";
import ErrorBoundary from "./components/ErrorBoundary";
import SearchHeader from "./components/search/SearchHeader";
import SearchPage from "./pages/SearchPage";
import SettingsPage from "./pages/SettingsPage";
import StatsPage from "./pages/StatsPage";
import InfoPage from "./pages/InfoPage";

function AppShell() {
  const location = useLocation();
  const isOther = ['/settings', '/stats', '/about', '/privacy'].includes(location.pathname);

  return (
    <LayoutShell
      header={!isOther ? <SearchHeader /> : <div />}
    >
      <Routes>
        <Route path="/" element={<SearchPage />} />
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="/stats" element={<StatsPage />} />
        <Route path="/about" element={<InfoPage />} />
        <Route path="/privacy" element={<InfoPage />} />
      </Routes>
    </LayoutShell>
  );
}

export default function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <AppShell />
      </BrowserRouter>
    </ErrorBoundary>
  );
}
