import React, { useEffect, useState } from 'react';
import { PluginCard } from './components/PluginCard';
import { getPlugins, updatePluginConfig, updatePluginStatus } from './services/api';
import { PluginWithStats } from './types/plugin';

function App() {
  const [plugins, setPlugins] = useState<PluginWithStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchPlugins = async () => {
    try {
      const data = await getPlugins();
      setPlugins(data);
      setError(null);
    } catch (err) {
      setError('Failed to load plugins. Please try again later.');
      console.error('Error fetching plugins:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPlugins();
    // Poll for updates every 5 seconds
    const interval = setInterval(fetchPlugins, 5000);
    return () => clearInterval(interval);
  }, []);

  const handleStatusChange = async (pluginName: string, isActive: boolean) => {
    try {
      await updatePluginStatus(pluginName, isActive);
      setPlugins((prev) =>
        prev.map((p) =>
          p.name === pluginName ? { ...p, isActive } : p
        )
      );
    } catch (err) {
      console.error('Error updating plugin status:', err);
    }
  };

  const handleConfigChange = async (pluginName: string, config: Record<string, any>) => {
    try {
      await updatePluginConfig(pluginName, config);
      setPlugins((prev) =>
        prev.map((p) =>
          p.name === pluginName ? { ...p, config } : p
        )
      );
    } catch (err) {
      console.error('Error updating plugin config:', err);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading plugins...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900">POS Plugin Management</h1>
          <p className="mt-2 text-gray-600">
            Manage and configure your Point-of-Sale plugins
          </p>
        </div>

        {error && (
          <div className="mb-8 bg-red-50 border border-red-200 rounded-md p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-red-400"
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-red-700">{error}</p>
              </div>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {plugins.map((plugin) => (
            <PluginCard
              key={plugin.name}
              plugin={plugin}
              onStatusChange={(isActive) => handleStatusChange(plugin.name, isActive)}
              onConfigChange={(config) => handleConfigChange(plugin.name, config)}
            />
          ))}
        </div>
      </div>
    </div>
  );
}

export default App; 
