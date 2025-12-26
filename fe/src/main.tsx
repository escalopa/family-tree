import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './i18n'; // Initialize i18n
import './styles/rtl.css'; // RTL/LTR support
import './styles/global-enhancements.css'; // Global UI enhancements
import './styles/colorful-enhancements.css'; // Colorful & engaging UI
import './styles/tree-branches-background.css'; // Tree branches decorative background

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
