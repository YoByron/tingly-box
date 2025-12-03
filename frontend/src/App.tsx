import CssBaseline from '@mui/material/CssBaseline';
import { ThemeProvider } from '@mui/material/styles';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';
import Dashboard from './pages/Dashboard';
import Login from './pages/Login';
import Providers from './pages/Providers';
import Server from './pages/Server';
import theme from './theme';

function App() {
    return (
        <ThemeProvider theme={theme}>
            <CssBaseline />
            <BrowserRouter>
                <AuthProvider>
                    <Routes>
                        <Route path="/login" element={<Login />} />
                        <Route
                            path="/*"
                            element={
                                <ProtectedRoute>
                                    <Layout>
                                        <Routes>
                                            <Route path="/" element={<Dashboard />} />
                                            <Route path="/dashboard" element={<Dashboard />} />
                                            <Route path="/providers" element={<Providers />} />
                                            <Route path="/server" element={<Server />} />
                                            {/* <Route path="/history" element={<History />} /> */}
                                        </Routes>
                                    </Layout>
                                </ProtectedRoute>
                            }
                        />
                    </Routes>
                </AuthProvider>
            </BrowserRouter>
        </ThemeProvider>
    );
}

export default App;
