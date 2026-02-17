import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ThemeProvider } from './context/ThemeContext';
import { Dashboard } from './pages/Dashboard';
import { ListView } from './pages/ListView';
import { Kanban } from './pages/Kanban';
import { SearchPage } from './pages/SearchPage';
import { Charts } from './pages/Charts';

function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/list" element={<ListView />} />
          <Route path="/kanban" element={<Kanban />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/charts" element={<Charts />} />
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;
